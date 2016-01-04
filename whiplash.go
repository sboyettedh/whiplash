package whiplash // github.com/sboyettedh/whiplash

import (
	"encoding/json"
)

const (
	// Version is the whiplash library version number.
	Version = "0.3.0"
)

// Row represents a datacenter row. It contains Racks.
type Row struct {
	ID int
	Name string
	Children []int
}

// Rack represents a datacenter rack. It contains Hosts.
type Rack struct {
	ID int
	Name string
	Children []int
}

// Host represents a machine running Ceph services. Svcs belong to it.
type Host struct {
	ID int
	Name string
	Rack string
}

// ClientUpdate is the struct used for interchange between whiplash clients
// and the aggregator. Each network request consists of the request
// name followed by whitespace followed by a JSON-encoded Request.
type ClientUpdate struct {
	// Time is the timestamp when the update was sent
	Time int64

	// Svc is the core identifying and status info about the service
	// making the request.
	Svc *SvcCore `json:"svc"`

	// Payload is the data accompanying the request. May be empty, as
	// in a ping request.
	Payload json.RawMessage `json:"payload"`
}

// QueryResponse is the standardized reply to a wlq query
type QueryResponse struct {
	// Code is like an HTTP code. 200 is success. 400 and up is error.
	Code int `json:"code"`
	// Cmd is the command which was dispatched
	Cmd string `json:"cmd"`
	// Subcmd is the dispatched subcommand
	Subcmd string `json:"subcmd"`
	// Args is the args which were passed in
	Args []string `json:"args"`
	// Data is the result of processing. It is generally the actual
	// response data, but if Code indicates an error, it will likely
	// be an error message.
	Data json.RawMessage `json:"data"`
}

// OsdStat is the data we want to ship to the aggregator about an OSD
// service's status
type OsdStat struct {
	// Weight is the crush weight of the OSD.
	Weight float32
	// BytesUsed is the amount of data stored on the OSD
	BytesUsed int
	// BytesAvail is the space remaining on the OSD
	BytesAvail int
	// PgPrimary is the number of PGs for which the OSD is the primary
	PgPrimary int
	// PgReplica is the number of PGs for which the OSD is a replica
	PgReplica int
}

// cephVersion represents the output of passing 'version' to a ceph admin daemon.
type cephVersion struct {
	Version string `json:"version"`
}

// cephOsdPerfDump represents the output of passing 'perf dump' to an
// OSD admin daemon.
type cephOsdPerfDump struct {
	WBThrottle json.RawMessage `json:"WBThrottle"`
	Filestore json.RawMessage `json:"filestore"`
	LevelDB json.RawMessage `json:"leveldb"`
	MutexFJCL json.RawMessage `json:"mutex-FileJournal::completions_lock"`
	MutexFJFL json.RawMessage `json:"mutex-FileJournal::finisher_lock"`
	MutexFJWL json.RawMessage `json:"mutex-FileJournal::write_lock"`
	MutexFHWQL json.RawMessage `json:"mutex-FileJournal::writeq_lock"`
	MutexJAMAL json.RawMessage `json:"mutex-JOS::ApplyManager::apply_lock"`
	MutexJAMCL json.RawMessage `json:"mutex-JOS::ApplyManager::com_lock"`
	MutexJSML json.RawMessage `json:"mutex-JOS::SubmitManager::lock"`
	MutexWBTL json.RawMessage `json:"mutex-WBThrottle::lock"`
	Objecter json.RawMessage `json:"objecter"`
	// Osd is currently the only thing we care about in here. It holds
	// tons of data about how the OSD is currently looking.
	Osd cephOsdPerfDumpOsd `json:"osd"`
	RecoveryState json.RawMessage `json:"recoverystate_perf"`
	ThrottleFSBytes json.RawMessage `json:"throttle-filestore_bytes"`
	ThrottleFSOps json.RawMessage `json:"throttle-filestore_ops"`
	ThrottleMDTCli json.RawMessage `json:"throttle-msgr_dispatch_throttler-client"`
	ThrottleMDTClu json.RawMessage `json:"throttle-msgr_dispatch_throttler-cluster"`
	ThrottleMDTHBBS json.RawMessage `json:"throttle-msgr_dispatch_throttler-hb_back_server"`
	ThrottleMDTHBFS json.RawMessage `json:"throttle-msgr_dispatch_throttler-hb_front_server"`
	ThrottleMDTHBC json.RawMessage `json:"throttle-msgr_dispatch_throttler-hbclient"`
	ThrottleMDTMO json.RawMessage `json:"throttle-msgr_dispatch_throttler-ms_objecter"`
	ThrottleObjBytes json.RawMessage `json:"throttle-objecter_bytes"`
	ThrottleObjOps json.RawMessage `json:"throttle-objecter_ops"`
	ThrottleOsdCB json.RawMessage `json:"throttle-osd_client_bytes"`
	ThrottleOsdCM json.RawMessage `json:"throttle-osd_client_messages"`
}

// cephOsdPerfDumpOsd is the "osd" section of the output of "perf dump"
type cephOsdPerfDumpOsd struct {
	OpWip int `json:"op_wip"`
	Op int `json:"op"`
	OpInBytes int `json:"op_in_bytes"`
	OpOutBytes int `json:"op_out_bytes"`
	OpLatency json.RawMessage `json:"op_latency"`
	OpProcessLatency json.RawMessage `json:"op_process_latency"`
	OpR int `json:"op_r"`
	OpROutBytes int `json:"op_r_out_bytes"`
	OpRLatency json.RawMessage `json:"op_r_latency"`
	OpRProcessLatency json.RawMessage `json:"op_r_process_latency"`
	OpW int `json:"op_w"`
	OpWInBytes int `json:"op_w_in_bytes"`
	OpWRLatency json.RawMessage `json:"op_w_rlat"`
	OpWLatency json.RawMessage `json:"op_w_latency"`
	OpWProcessLatency json.RawMessage `json:"op_w_process_latency"`
	OpRW int `json:"op_rw"`
	OpRWInBytes int `json:"op_rw_in_bytes"`
	OpRWOutBytes int `json:"op_rw_out_bytes"`
	OpRWRLatency json.RawMessage `json:"op_rw_rlat"`
	OpRWLatency json.RawMessage `json:"op_rw_latency"`
	OpRWProcessLatency json.RawMessage `json:"op_rw_process_latency"`
	Subop int `json:"subop"`
	SubopInBytes int `json:"subop_in_bytes"`
	SubopLatency json.RawMessage `json:"subop_latency"`
	SubopW int `json:"subop_w"`
	SubopWInBytes int `json:"subop_w_in_bytes"`
	SubopWLatency json.RawMessage `json:"subop_w_latency"`
	SubopPull int `json:"subop_pull"`
	SubopPullLatency json.RawMessage `json:"subop_pull_latency"`
	SubopPush int `json:"subop_push"`
	SubopPushInBytes int `json:"subop_push_in_bytes"`
	SubopPushLatency json.RawMessage `json:"subop_push_latency"`
	Pull int `json:"pull"`
	Push int `json:"push"`
	PushOutBytes int `json:"push_out_bytes"`
	PushIn int `json:"push_in"`
	PushInBytes int `json:"push_in_bytes"`
	RecoveryOps int `json:"recovery_ops"`
	Loadavg int `json:"loadavg"`
	BufferBytes int `json:"buffer_bytes"`
	// NumPg is the count of PGs on this OSD
	NumPg int `json:"numpg"`
	// NumPgPrimary is the count of PGs for which this OSD is the primary
	NumPgPrimary int `json:"numpg_primary"`
	// NumPgReplica is the count of PGs for which this OSD is a replica
	NumPgReplica int `json:"numpg_replica"`
	NumPgStray int `json:"numpg_stray"`
	HeartbeatToPeers int `json:"heartbeat_to_peers"`
	HeartbeatFromPeers int `json:"heartbeat_from_peers"`
	MapMsgs int `json:"map_messages"`
	MapMsgEpochs int `json:"map_message_epochs"`
	MapMsgEpochDups int `json:"map_message_epoch_dups"`
	MsgsDelayedForMap int `json:"messages_delayed_for_map"`
	// StatBytes is the total size of the OSDs filesystem
	StatBytes int `json:"stat_bytes"`
	// StatBytesUsed is the amount of data stored by the OSD
	StatBytesUsed int `json:"stat_bytes_used"`
	// StatBytesAvail is the space remaining on the OSD
	StatBytesAvail int `json:"stat_bytes_avail"`
	Copyfrom int `json:"copyfrom"`
	TierPromote int `json:"tier_promote"`
	TierFlush int `json:"tier_flush"`
	TierFlushFail int `json:"tier_flush_fail"`
	TierTryFlush int `json:"tier_try_flush"`
	TierTryFlushFail int `json:"tier_try_flush_fail"`
	TierEvict int `json:"tier_evict"`
	TierWhiteout int `json:"tier_whiteout"`
	TierDirty int `json:"tier_dirty"`
	TierClean int `json:"tier_clean"`
	TierDelay int `json:"tier_delay"`
	TierProxyDead int `json:"tier_proxy_read"`
	AgentWake int `json:"agent_wake"`
	AgentSkip int `json:"agent_skip"`
	AgentFlush int `json:"agent_flush"`
	AgentEvict int `json:"agent_evict"`
	ObjCtxCacheHit int `json:"object_ctx_cache_hit"`
	ObjCtxCacheTotal int `json:"object_ctx_cache_total"`
}
