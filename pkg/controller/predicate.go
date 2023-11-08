/*
Copyright (c) 2023 OceanBase
ob-operator is licensed under Mulan PSL v2.
You can use this software according to the terms and conditions of the Mulan PSL v2.
You may obtain a copy of Mulan PSL v2 at:
         http://license.coscl.org.cn/MulanPSL2
THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
See the Mulan PSL v2 for more details.
*/

package controller

import (
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

var preds = predicate.GenerationChangedPredicate{}

// var preds = predicate.Or(
// 	predicate.AnnotationChangedPredicate{},
// 	predicate.LabelChangedPredicate{},
// 	predicate.GenerationChangedPredicate{},
// 	predicate.Funcs{
//		// Default value of funcs is true
// 		CreateFunc: func(event.CreateEvent) bool {
// 			return false
// 		},
// 		DeleteFunc: func(event.DeleteEvent) bool {
// 			return true
// 		},
// 		UpdateFunc: func(event.UpdateEvent) bool {
// 			return false
// 		},
// 		GenericFunc: func(event.GenericEvent) bool {
// 			return false
// 		},
// 	},
// )
