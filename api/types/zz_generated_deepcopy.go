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

package types

func (in *LogServiceZoneTopology) DeepCopyInto(out *LogServiceZoneTopology) {
	*out = *in
	if in.NodeSelector != nil {
		in, out := &in.NodeSelector, &out.NodeSelector
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
}

func (in *LogServiceZoneTopology) DeepCopy() *LogServiceZoneTopology {
	if in == nil {
		return nil
	}
	out := new(LogServiceZoneTopology)
	in.DeepCopyInto(out)
	return out
}

func (in *LogServiceStorageSpec) DeepCopyInto(out *LogServiceStorageSpec) {
	*out = *in
	if in.StoreStorage != nil {
		in, out := &in.StoreStorage, &out.StoreStorage
		*out = new(StorageSpec)
		**out = **in
	}
	if in.LogStorage != nil {
		in, out := &in.LogStorage, &out.LogStorage
		*out = new(StorageSpec)
		**out = **in
	}
}

func (in *LogServiceStorageSpec) DeepCopy() *LogServiceStorageSpec {
	if in == nil {
		return nil
	}
	out := new(LogServiceStorageSpec)
	in.DeepCopyInto(out)
	return out
}

func (in *ObjectStoreConfig) DeepCopyInto(out *ObjectStoreConfig) {
	*out = *in
}

func (in *ObjectStoreConfig) DeepCopy() *ObjectStoreConfig {
	if in == nil {
		return nil
	}
	out := new(ObjectStoreConfig)
	in.DeepCopyInto(out)
	return out
}

func (in *LogServiceZoneReplicaStatus) DeepCopyInto(out *LogServiceZoneReplicaStatus) {
	*out = *in
}

func (in *LogServiceZoneReplicaStatus) DeepCopy() *LogServiceZoneReplicaStatus {
	if in == nil {
		return nil
	}
	out := new(LogServiceZoneReplicaStatus)
	in.DeepCopyInto(out)
	return out
}

func (in *LogServiceNodeReplicaStatus) DeepCopyInto(out *LogServiceNodeReplicaStatus) {
	*out = *in
}

func (in *LogServiceNodeReplicaStatus) DeepCopy() *LogServiceNodeReplicaStatus {
	if in == nil {
		return nil
	}
	out := new(LogServiceNodeReplicaStatus)
	in.DeepCopyInto(out)
	return out
}

func (in *SharedStorageSpec) DeepCopyInto(out *SharedStorageSpec) {
	*out = *in
}

func (in *SharedStorageSpec) DeepCopy() *SharedStorageSpec {
	if in == nil {
		return nil
	}
	out := new(SharedStorageSpec)
	in.DeepCopyInto(out)
	return out
}

func (in *LogServiceReference) DeepCopyInto(out *LogServiceReference) {
	*out = *in
}

func (in *LogServiceReference) DeepCopy() *LogServiceReference {
	if in == nil {
		return nil
	}
	out := new(LogServiceReference)
	in.DeepCopyInto(out)
	return out
}
