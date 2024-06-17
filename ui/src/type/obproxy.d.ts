declare namespace OBProxy {
  interface CommonProxyDetail {
    name: string;
    namespace: string;
    image?: string;
    parameters?: CommonKVPair[];
    resource?: CommonResourceSpec;
    serviceType?: string;
    replicas?: number;
  }
}

export { OBProxy };
