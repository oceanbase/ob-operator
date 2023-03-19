/*
Copyright (c) 2021 OceanBase
ob-operator is licensed under Mulan PSL v2.
You can use this software according to the terms and conditions of the Mulan PSL v2.
You may obtain a copy of Mulan PSL v2 at:
         http://license.coscl.org.cn/MulanPSL2
THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
See the Mulan PSL v2 for more details.
*/

package core

import (
	"context"
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"

	cloudv1 "github.com/oceanbase/ob-operator/apis/cloud/v1"
	backupconst "github.com/oceanbase/ob-operator/pkg/controllers/backup/const"
	"github.com/oceanbase/ob-operator/pkg/controllers/backup/sql"
	observerconst "github.com/oceanbase/ob-operator/pkg/controllers/observer/const"
	"github.com/oceanbase/ob-operator/pkg/controllers/observer/core/converter"
	"github.com/oceanbase/ob-operator/pkg/infrastructure/kube/resource"
	"github.com/oceanbase/ob-operator/pkg/util"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	kubeerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// BackupReconciler reconciles a backup object
type BackupReconciler struct {
	CRClient client.Client
	Scheme   *runtime.Scheme

	Recorder record.EventRecorder
}

type BackupCtrl struct {
	Resource *resource.Resource
	Backup   cloudv1.Backup
}

type BackupCtrlOperator interface {
	BackupCoordinator() (ctrl.Result, error)
}

// +kubebuilder:rbac:groups=cloud.oceanbase.com,resources=backups,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cloud.oceanbase.com,resources=backups/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=cloud.oceanbase.com,resources=backups/finalizers,verbs=update
// +kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=services/status,verbs=get;update;patch
// +kubebuilder:rbac:groups="",resources=services/finalizers,verbs=update
// +kubebuilder:rbac:groups="",resources=events,verbs=create;patch

func (r *BackupReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	// Fetch the CR instance
	instance := &cloudv1.Backup{}
	err := r.CRClient.Get(ctx, req.NamespacedName, instance)
	if err != nil {
		if kubeerrors.IsNotFound(err) {
			// Object not found, return.
			// Created objects are automatically garbage collected.
			return reconcile.Result{}, nil
		}
		// Error reading the object, requeue the request.
		return reconcile.Result{}, err
	}
	// custom logic
	backupCtrl := NewBackupCtrl(r.CRClient, r.Recorder, *instance)

	backupFinalizerName := fmt.Sprintf("cloud.oceanbase.com.finalizers.%s", instance.Name)
	if instance.ObjectMeta.DeletionTimestamp.IsZero() {
		if !util.ContainsString(instance.ObjectMeta.Finalizers, backupFinalizerName) {
			instance.ObjectMeta.Finalizers = append(instance.ObjectMeta.Finalizers, backupFinalizerName)
			if err := r.CRClient.Update(context.Background(), instance); err != nil {
				return ctrl.Result{}, err
			}
		}
	} else {
		if util.ContainsString(instance.ObjectMeta.Finalizers, backupFinalizerName) {
			err := r.BackupDelete(ctx, r.CRClient, r.Recorder, instance)
			if err != nil {
				return ctrl.Result{}, err
			}
			instance.ObjectMeta.Finalizers = util.RemoveString(instance.ObjectMeta.Finalizers, backupFinalizerName)
			if err := r.CRClient.Update(context.Background(), instance); err != nil {
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, nil
	}

	// Fetch the OBCluster CR instance
	obNamespace := types.NamespacedName{
		Namespace: instance.Spec.SourceCluster.ClusterNamespace,
		Name:      instance.Spec.SourceCluster.ClusterName,
	}
	obInstance := &cloudv1.OBCluster{}
	err = r.CRClient.Get(ctx, obNamespace, obInstance)
	if err != nil {
		if kubeerrors.IsNotFound(err) {
			klog.Infof("OBCluster %s not found, namespace %s", instance.Spec.SourceCluster.ClusterName, instance.Spec.SourceCluster.ClusterNamespace)
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}
	if obInstance.Status.Status != observerconst.ClusterReady {
		klog.Infoln("OBCluster  %s is not ready, namespace %s", instance.Spec.SourceCluster.ClusterName, instance.Spec.SourceCluster.ClusterNamespace)
		return reconcile.Result{}, nil
	}

	return backupCtrl.BackupCoordinator()
}

func (r *BackupReconciler) BackupDelete(ctx context.Context, client client.Client, recorder record.EventRecorder, backup *cloudv1.Backup) error {
	ctrlResource := resource.NewResource(client, recorder)
	ctrl := &BackupCtrl{
		Backup:   *backup,
		Resource: ctrlResource,
	}
	obNamespace := types.NamespacedName{
		Namespace: backup.Namespace,
		Name:      backup.Spec.SourceCluster.ClusterName,
	}
	obInstance := &cloudv1.OBCluster{}
	err := r.CRClient.Get(ctx, obNamespace, obInstance)
	if err != nil {
		if kubeerrors.IsNotFound(err) {
			return nil
		}
		return err
	}
	isArchivelogStop, err := ctrl.IsArchivelogStop()
	if err != nil {
		klog.Errorf("check archivelog stop, error '%s'", err)
		return err
	}
	if !isArchivelogStop {
		err = ctrl.CancelArchiveLog()
		if err != nil {
			klog.Errorf("cancel archivelog failed, error '%s'", err)
			return err
		}
	}
	return nil
}

func NewBackupCtrl(client client.Client, recorder record.EventRecorder, backup cloudv1.Backup) BackupCtrlOperator {
	ctrlResource := resource.NewResource(client, recorder)
	return &BackupCtrl{
		Resource: ctrlResource,
		Backup:   backup,
	}
}

func (ctrl *BackupCtrl) BackupCoordinator() (ctrl.Result, error) {
	// Backup control-plan
	err := ctrl.BackupEffector()
	if err != nil {
		return reconcile.Result{}, err
	}
	return reconcile.Result{}, nil
}

func (ctrl *BackupCtrl) BackupEffector() error {
	backupSets := ctrl.Backup.Status.BackupSet
	isExist := false
	var existBackupType string
	for _, backupSet := range backupSets {
		if backupSet.ClusterName == ctrl.Backup.Spec.SourceCluster.ClusterName {
			isExist = true
			if backupSet.BackupType == backupconst.DatabaseBackupType {
				existBackupType = backupconst.FullBackup
			}
			if backupSet.BackupType == backupconst.IncDatabaseBackupType {
				existBackupType = backupconst.IncrementalBackup
			}
			break
		}
	}
	if isExist {
		for _, schedule := range ctrl.Backup.Spec.Schedule {
			if schedule.Schedule == backupconst.BackupOnce && schedule.BackupType == existBackupType {
				return nil
			}
		}
	}
	return ctrl.DoBackup()
}

func (ctrl *BackupCtrl) DoBackup() error {
	err, isBackupDestSet := ctrl.isBackupDestSet()
	if err != nil {
		klog.Errorln("DoBackup: check whether Backup Dest is Set err ", err)
		return err
	}
	if !isBackupDestSet {
		dest_path := ctrl.Backup.Spec.DestPath
		return ctrl.SetBackupDest(dest_path)
	}

	err = ctrl.setBackupDestOption()
	if err != nil {
		klog.Errorln("DoBackup: set BackupDest Option err ", err)
		return err
	}

	archiveLogStatus, err := ctrl.GetArchivelogStatus()
	if err != nil {
		klog.Errorln("DoBackup: get Archivelog status err ", err)
		return err
	}
	switch archiveLogStatus {
	case backupconst.ArchiveLogInterrupted:
		klog.Errorln("archivelog status is interrupted")
		return errors.New("archivelog status is interrupted")
	case backupconst.ArchiveLogDoing:
		break
	case backupconst.ArchiveLogPrepare, backupconst.ArchiveLogBeginning:
		err = ctrl.WaitArchivelogDoing()
		if err != nil {
			klog.Errorln("wait backup logArchive doing err ", err)
			return err
		}
	case backupconst.ArchiveLogStopping:
		klog.Infoln("archivelog status is stopping")
		return nil
	case backupconst.ArchiveLogStop:
		err = ctrl.setBackupLogArchiveOption()
		if err != nil {
			klog.Errorln("DoBackup: set Backup LogArchive Option err ", err)
			return err
		}
		err = ctrl.setBackupLogArchive()
		if err != nil {
			klog.Errorln("DoBackup: set Backup LogArchive err ", err)
			return err
		}
		err = ctrl.WaitArchivelogDoing()
		if err != nil {
			klog.Errorln("wait backup logArchive doing err ", err)
			return err
		}

	}

	for _, schedule := range ctrl.Backup.Spec.Schedule {
		err, isBackupDoing := ctrl.isBackupDoing()
		if err != nil {
			klog.Errorln("DoBackup: check whether backup is doing err ", err)
			return err
		}
		if isBackupDoing {
			continue
		}
		// deal with full backup
		if schedule.BackupType == backupconst.FullBackup {
			// full backup once
			if schedule.Schedule == backupconst.BackupOnce {
				err = ctrl.StartBackupDatabase()
				if err != nil {
					klog.Errorln("DoBackup: Start Backup Database err ", err)
					return err
				}
				return ctrl.UpdateBackupStatus("")
				//full backup, periodic
			} else {
				scheduleStatus := ctrl.getBackupScheduleStatus(backupconst.FullBackupType)
				// first time
				if scheduleStatus.NextTime == "" {
					return ctrl.UpdateBackupStatus("")
				}
				nextTime, err := time.ParseInLocation("2006-01-02 15:04:05 +0800 CST", scheduleStatus.NextTime, time.Local)
				if err != nil {
					klog.Errorln("DoBackup: full backup time Parse err ", err)
					return err
				}
				if nextTime.Before(time.Now()) || nextTime.Equal(time.Now()) {
					err = ctrl.StartBackupDatabase()
					if err != nil {
						klog.Errorln("DoBackup: full backup Start Backup Database err ", err)
						return err
					}
					return ctrl.UpdateBackupStatus(backupconst.FullBackupType)
				}
			}

		}
		// deal with incremental backup
		if schedule.BackupType == backupconst.IncrementalBackup {
			// incremental backup once
			if schedule.Schedule == backupconst.BackupOnce {
				err = ctrl.StartBackupIncremental()
				if err != nil {
					klog.Errorln("DoBackup: Start Backup Incremental err ", err)
					return err
				}
				// incremental backup, periodic
			} else {
				scheduleStatus := ctrl.getBackupScheduleStatus(backupconst.IncrementalBackupType)
				// first time
				if scheduleStatus.NextTime == "" {
					return ctrl.UpdateBackupStatus("")
				}
				nextTime, err := time.ParseInLocation("2006-01-02 15:04:05 +0800 CST", scheduleStatus.NextTime, time.Local)
				if err != nil {
					klog.Errorln("DoBackup: Incremental backup time Parse err ", err)
					return err
				}
				if nextTime.Before(time.Now()) || nextTime.Equal(time.Now()) {
					err = ctrl.StartBackupIncremental()
					if err != nil {
						klog.Errorln("DoBackup: Incremental Backup Start Backup Incremental err ", err)
						return err
					}
					return ctrl.UpdateBackupStatus(backupconst.IncrementalBackupType)
				}
			}
		}
	}
	return ctrl.UpdateBackupStatus("")
}

func (ctrl *BackupCtrl) GetSqlOperator() (*sql.SqlOperator, error) {
	clusterIP, err := ctrl.GetServiceClusterIPByName(ctrl.Backup.Namespace, ctrl.Backup.Spec.SourceCluster.ClusterName)
	// get svc failed
	if err != nil {
		return nil, errors.New("failed to get service address")
	}
	secretName := converter.GenerateSecretNameForDBUser(ctrl.Backup.Spec.SourceCluster.ClusterName, "sys", "admin")
	secretExecutor := resource.NewSecretResource(ctrl.Resource)
	secret, err := secretExecutor.Get(context.TODO(), ctrl.Backup.Namespace, secretName)
	user := "root"
	password := ""
	if err == nil {
		user = "admin"
		password = string(secret.(corev1.Secret).Data["password"])
	}

	p := &sql.DBConnectProperties{
		IP:       clusterIP,
		Port:     observerconst.MysqlPort,
		User:     user,
		Password: password,
		Database: "oceanbase",
		Timeout:  10,
	}
	so := sql.NewSqlOperator(p)
	if so.TestOK() {
		return so, nil
	}
	return nil, errors.New("failed to get sql operator")
}

func (ctrl *BackupCtrl) GetServiceClusterIPByName(namespace, name string) (string, error) {
	svcName := converter.GenerateServiceName(name)
	serviceExecuter := resource.NewServiceResource(ctrl.Resource)
	svc, err := serviceExecuter.Get(context.TODO(), namespace, svcName)
	if err != nil {
		return "", err
	}
	return svc.(corev1.Service).Spec.ClusterIP, nil
}
