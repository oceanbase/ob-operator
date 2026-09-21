// Current Dashboard creation contract, verified against oblogservice:1.3.0.
// Running clusters may scale beyond this initial bootstrap topology.
export function validBootstrapReplicas(zones?: { replica: number }[]): boolean {
  return (
    !!zones?.length &&
    zones.every((zone) => Number.isInteger(zone.replica) && zone.replica > 0) &&
    zones.reduce((total, zone) => total + zone.replica, 0) === 3
  );
}
