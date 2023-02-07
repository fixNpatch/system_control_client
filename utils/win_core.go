package utils

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

var systemHosts = os.Getenv("SystemRoot") + "/System32/drivers/etc/hosts"

func DoProcess() bool {
	success := true
	//hostsIPMap := getHostsIpMap()
	//overwriteSystemHosts()
	//processNameMap := GetProcessNameMap()
	table := getTCPTable()
	for i := uint32(0); i < uint32(table.dwNumEntries); i++ {
		row := table.table[i]
		ip := row.displayIP(row.dwRemoteAddr)
		port := row.displayPort(row.dwRemotePort)

		fmt.Println(ip, port)

		if row.dwOwningPid <= 0 {
			continue
		}
		if port != 80 && port != 443 {
			continue
		}

		//process := strings.ToLower(processNameMap[uint32(row.dwOwningPid)])
		//if hostsIPMap[ip] {
		//	if err := CloseTCPEntry(row); err != nil {
		//		fmt.Printf("Fail to close TCP connections: Process = %v, Pid = %v, Addr = %v:%v", process, row.dwOwningPid, ip, port)
		//		success = false
		//	} else {
		//		fmt.Printf("Succeed to close TCP connections: Process = %v, Pid = %v, Addr = %v:%v", process, row.dwOwningPid, ip, port)
		//	}
		//}
	}
	return success
}

// find all the ip of system and current hosts
//func getHostsIpMap() map[string]bool {
//	ipMap := make(map[string]bool)
//	for _, v := range readHostConfigMap(systemHosts) {
//		ipMap[v] = true
//	}
//	model := conf.Config.HostConfigModel
//	index := conf.Config.CurrentHostIndex
//	if length := len(model.Roots); index >= 0 && length > 0 && index < length {
//		path := "conf/hosts/" + model.RootAt(index).Text() + ".hosts"
//		for _, v := range readHostConfigMap(path) {
//			ipMap[v] = true
//		}
//	}
//	return ipMap
//}
//
//func overwriteSystemHosts() {
//	bytes := ReadCurrentHostConfig()
//	if bytes == nil {
//		return
//	}
//	if err := ioutil.WriteFile(systemHosts, bytes, os.ModeExclusive); err != nil {
//		fmt.Printf("Error writing to system hosts file: %v", err)
//	}
//}
//
//func ReadCurrentHostConfig() []byte {
//	model := conf.Config.HostConfigModel
//	index := conf.Config.CurrentHostIndex
//	if length := len(model.Roots); index < 0 || length <= 0 || index >= length {
//		return nil
//	}
//	bytes, err := ioutil.ReadFile("conf/hosts/" + model.RootAt(index).Text() + ".hosts")
//	if err != nil {
//		fmt.Printf("Error reading host config: %v", err)
//		return nil
//	}
//	return bytes
//}

func readHostConfigMap(path string) map[string]string {
	hostConfigMap := make(map[string]string)
	file, err := os.Open(path)
	if err != nil {
		fmt.Printf("Fail to open system_hosts: %s", err)
		return hostConfigMap
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && !strings.HasPrefix(line, "#") {
			config := strings.Fields(line)
			if len(config) == 2 {
				hostConfigMap[config[1]] = config[0]
			}
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Printf("Fail to read system_hosts: %s", err)
	}
	return hostConfigMap
}
