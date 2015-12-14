package whiplash

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

// WLConfig is the overall Whiplash configuration.
type WLConfig struct {
	// Aggregator config
	Aggregator WLAggConfig `json:"aggregator"`

	// Agent config
	Agent WLAgtConfig `json:"agent"`

	// Location of the Ceph configuration file
	CephConfLoc string `json:"cephconf_loc"`

	// Parsed Ceph configuration
	CephConf map[string]map[string]string `json:"ceph_conf"`

	// Ceph services discovered via the config file
	Svcs map[string]*Svc `json:"services"`
}

// WLAggConfig is the Whiplash aggregator configuration.
type WLAggConfig struct {
	// BindAddr is the address the aggregator should bind to.
	BindAddr string `json:"bind_addr"`
	// BindPort is the port to bind to on BindAddr
	BindPort string `json:"bind_port"`
	// MsgLvl sets the Asock.Config Msglvl parameter. Valid: "all",
	// "conn", "error", "fatal".
	MsgLvl string `json:"msglvl"`
}

// WLAgtConfig is the Whiplash agent configuration.
type WLAgtConfig struct {
}

// NewConfig returns a populated Whiplash configuration. `wlconf` is
// the whiplash configuration file. `gensvcs` controls whether an
// attempt will be made to read the Ceph configuration and construct a
// list of services active on the machine.
func New(wlconf string, gensvcs bool) (*WLConfig, error) {
	wlc := &WLConfig{}
	err := wlc.getConfig(wlconf, gensvcs)
	if err != nil {
		return nil, err
	}
	return wlc, err
}

// getConfig reads the specified whiplash config file and returns its
// contents, along with the contents of the Ceph configuration file it
// points to.
func (wlc *WLConfig) getConfig(wlconf string, gensvcs bool) (error) {
	conffile, err := ioutil.ReadFile(wlconf)
	if err != nil {
		return err
	}
	err = json.Unmarshal(conffile, wlc)
	if err != nil {
		return err
	}
	if wlc.CephConfLoc == "" {
		return fmt.Errorf("No Ceph configuration found in `%v`", wlc.CephConfLoc)
	}
	if wlc.Aggregator.BindAddr == "" {
		return fmt.Errorf("No aggregator address found in `%v`", wlc.CephConfLoc)
	}
	ml := wlc.Aggregator.MsgLvl
	if ml != "all" && ml != "conn" && ml != "error" && ml != "fatal" {
		return fmt.Errorf("Aggregator.Msglvl must be one of 'all', 'conn', 'error', 'fatal'")
	}
	if gensvcs {
		wlc.CephConf, err = parseCephConf(wlc.CephConfLoc)
		if err != nil {
			return err
		}
		wlc.getCephServices()
	}
	return nil
}

// parseCephConf reads a Ceph configuration file, and turns it into a
// map of maps. The top-level map has the conf file section names as
// keys. The second-level maps contain the key-value pairs of each
// section of the configuration.
func parseCephConf(cephconf string) (map[string]map[string]string, error) {
	conffile, err := os.Open(cephconf)
	if err != nil {
		return nil, err
	}
	defer conffile.Close()
	cm := make(map[string]map[string]string)
	confreader := bufio.NewReader(conffile)
	confscanner := bufio.NewScanner(confreader)
	section := ""
	for confscanner.Scan() {
		line := confscanner.Text()
		line = strings.TrimSpace(line)
		// skip comments
		if strings.HasPrefix(line, "#") {
			continue
		}
		// handle section markers
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			section = strings.Trim(line, "[]")
			cm[section] = map[string]string{}
			continue
		}
		// handle regular lines
		chunks := strings.Split(line, " = ")
		if len(chunks) < 2 {
			chunks = strings.Split(line, "=")
		}
		if len(chunks) < 2 {
			continue
		}
		cm[section][chunks[0]] = chunks[1]
	}
	return cm, nil
}
