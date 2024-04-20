/*
Management of the connections between replicas and clients.
Instead of waiting for the timeout of a connection,
we take a proactive approach where a node is put in 'blacklist' if it cannot be connected.
The nodes that are in blacklist will be moved out of the list if they join the system.
*/

/*TODO:
The current version simply 'Blacklist' nodes that have connection error and puts nodes back if they join the system again.
We need a scheme to optimize this either via
1) enhancing the implementation to avoid slowdown
or
2) put node back to the list when a receiver is reachable again.
Current version puts a node back upon a join request (by the same node)

In the future version, we can integrate this with recovery module
*/

package communication

import (
	"chainhealth/src/utils"
)

var connection utils.StringBoolMap

/*
Check whether a not is not alive
Input

	key: string format, a node id

Output

	bool: whether the node is set to not alive, true if the node is not alive, false otherwise.
*/
func IsNotLive(key string) bool {
	result, exist := connection.Get(key)
	if exist {
		return result
	}
	return false
}

/*
Set node to not alive
Input

	key: string format, a node id
*/
func NotLive(key string) {
	connection.Insert(key, true)
}

/*
Set node to live
Input

	key: string format, a node id
*/
func IsLive(key string) bool {
	result, exist := connection.Get(key)
	if exist && result == false {
		return true
	}
	return false
}

/*
Set node to live
Input

	key: string format, a node id
*/
func SetLive(key string) {
	connection.Insert(key, false)
}

/*
Start connection manager
*/
func StartConnectionManager() {
	connection.Init()
}
