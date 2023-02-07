package main

import (
	"diplom_client/structs"
	"diplom_client/utils"
	"encoding/json"
	"fmt"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/mem"
	"strconv"

	//"strconv"
	"time"
	//"github.com/shirou/gopsutil/process"
	"golang.org/x/net/websocket"
	"net/http"
)

var globalStorage interface{}

func getStats() (*structs.Stats, error) {
	vmStat, err := mem.VirtualMemory()
	if err != nil {
		fmt.Println("VirtualMemory error: ", err)
		return nil, err
	}

	diskStat, err := disk.Usage("/")
	if err != nil {
		fmt.Println("Disk.Usage error: ", err)
		return nil, err
	}
	cpuStats, err := cpu.Info()
	if err != nil {
		fmt.Println("Cpu.Info error: ", err)
		return nil, err
	}
	percentage, err := cpu.Percent(0, true)
	if err != nil {
		fmt.Println("Cpu.Percent error: ", err)
		return nil, err
	}
	hostStat, err := host.Info()
	if err != nil {
		fmt.Println("Host.Info error: ", err)
		return nil, err
	}

	//procStat, err := process.Processes()
	//if err != nil {
	//	fmt.Println("Proc.List error: ", err)
	//	return nil, err
	//}

	connections, err := utils.GetExtendedTCPTable()
	if err != nil {
		fmt.Println("conn error: ", err)
		return nil, err
	}
	stats := structs.Stats{
		VmStat: structs.VmStat{
			Total:       vmStat.Total,
			Available:   vmStat.Available,
			Used:        vmStat.Used,
			UsedPercent: vmStat.UsedPercent,
			Free:        vmStat.Free,
			Active:      vmStat.Active,
			Inactive:    vmStat.Inactive,
			Wired:       vmStat.Wired,
		},
		//Processes: structs.Procs{Pid: procStat},
		Disk: structs.Disk{
			Total:       diskStat.Total,
			Used:        diskStat.Used,
			Free:        diskStat.Free,
			UsedPercent: diskStat.UsedPercent,
		},
		Cpu: structs.Cpu{
			Percentage: percentage,
			Model:      cpuStats[0].ModelName,
			Cores:      int(cpuStats[0].Cores),
		},
		Host: structs.HostInfo{
			Procs:           hostStat.Procs,
			OS:              hostStat.OS,
			Platform:        hostStat.Platform,
			PlatformVersion: hostStat.PlatformVersion,
		},
		Connections: connections,
		HostTime:    time.Now(),
	}

	return &stats, err
}

type wsAuth struct {
	login    string
	password string
}

func auth(ws *websocket.Conn) {
	fmt.Println("auth")
	var data wsAuth
	if err := websocket.JSON.Receive(ws, &data); err != nil {
		fmt.Println(err)
	}
	fmt.Println(data)
}

func echo(ws *websocket.Conn) {
	fmt.Println("echo")

	var cmd string
	var data interface{}
	var err error

	for {

		if err := websocket.Message.Receive(ws, &cmd); err != nil {
			fmt.Println(err)
			return
		}

		switch cmd {
		case "fupd": // full update
			data, err = getStats()
			if err != nil {
				fmt.Println("Error with getting stats::", err)
				err = ws.Close()
				if err != nil {
					fmt.Println("Error with closing websocket::", err)
				}
			}
			break
		case "pupd":
			break
		}

		if err = websocket.JSON.Send(ws, &data); err != nil {
			fmt.Println(err)
			err = ws.Close()
			if err != nil {
				fmt.Println(err)
			}
		}
	}
}

func simple(w http.ResponseWriter, r *http.Request) {

}

func router(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/":
		c, err := r.Cookie("auth")
		if err != nil {
			fmt.Println("Case auth::", err)
			//return
		} else if c.Value == "1" {
			//break
			fmt.Println("Case auth:: auth cookie is 1")
		}
		websocket.Server{Handler: websocket.Handler(echo)}.ServeHTTP(w, r)
		break
	case "/close":

		fmt.Println("call close")
		fmt.Println(r)

		var params map[string]interface{}

		err := json.NewDecoder(r.Body).Decode(&params)
		if err != nil {
			fmt.Println(err)
		}

		tmp := make(map[string]string, len(params))

		for i := range params {
			tmp[i] = fmt.Sprintf("%v", params[i])
		}

		isOk := closeTCPConnection(tmp)

		// TODO сделать правильную отправку ответа

		if isOk {
			if _, err := w.Write([]byte("ok")); err != nil {
				fmt.Println(err)
			}
		} else {
			if _, err := w.Write([]byte("err")); err != nil {
				fmt.Println(err)
			}
		}

		//keys, err := url.ParseQuery(r.URL.RawQuery)
		//if err != nil {
		//	fmt.Println(err)
		//}
		//closeTCPConnection(keys["pid"][0])
	case "/auth":
		cookie := http.Cookie{Name: "auth", Value: "1", Path: "/", HttpOnly: true, Secure: false}

		server := websocket.Server{Handler: websocket.Handler(auth)}
		server.Config.Header = make(map[string][]string)
		server.Config.Header.Set("Set-Cookie", cookie.String())
		server.ServeHTTP(w, r)

		//websocket.Server{Handler: websocket.Handler(auth)}.ServeHTTP(w, r)

		//http.S{Handler:http.Handler(simple)}.Se

		break

	}

	//fmt.Println(r.URL.Path)

	//server := websocket.Server{Handler: websocket.Handler(echo)}
	//server.ServeHTTP(w, r)
}

func closeTCPConnection(params map[string]string) (ok bool) {

	fmt.Println(params)

	pid, err := strconv.ParseInt(params["pid"], 10, 64)
	if err != nil {
		fmt.Println("closeTCPConnection::", err)
	}

	rport, err := strconv.ParseInt(params["rport"], 10, 64)
	if err != nil {
		fmt.Println("closeTCPConnection::", err)
	}

	lport, err := strconv.ParseInt(params["lport"], 10, 64)
	if err != nil {
		fmt.Println("closeTCPConnection::", err)
	}

	if _, found := params["laddr"]; !found {
		fmt.Println("laddr param not found")
		return
	}
	laddr := params["laddr"]

	if _, found := params["raddr"]; !found {
		fmt.Println("raddr param not found")
		return
	}
	raddr := params["raddr"]

	return utils.CloseConnection(pid, laddr, lport, raddr, rport) // windows only
}

// точка старта программы
func main() {
	http.HandleFunc("/", router)             // обработчик входящих запросов
	err := http.ListenAndServe(":9010", nil) // обработчик входящих подключений
	if err != nil {
		fmt.Println("Cannot start server. Reason: ", err.Error())
		return
	}
}
