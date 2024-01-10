import { useRequest } from "ahooks"
import { getTenant } from "@/services/tenant"
// 租户概览页
export default function TenantPage(){

    // const {run:_getTenant}  = useRequest(getTenant,{
    //     onSuccess:(res)=>{
    //         res.data
    //     }
    // })

    return <h1>TenantDetail</h1>
}