package main

import (
	"C"
	"fmt"
	"log"
	"os/exec"
	"syscall"
	"time"
	"unsafe"

	"github.com/go-co-op/gocron"
	nvml "github.com/mxpv/nvml-go"
	// "winbase_exporter/nvml"
)

type Device uintptr

const TemperatureGPU = nvml.TemperatureSensor(0)

// 获取显卡温度
func GetTemp() uint {
	nvml := syscall.NewLazyDLL("nvml.dll")
	initProc := nvml.NewProc("nvmlInit")
	_, _, _ = initProc.Call()
	//通过下标索引来获取显卡设备的句柄
	var device uintptr
	deviceProc := nvml.NewProc("nvmlDeviceGetHandleByIndex")
	_, _, _ = deviceProc.Call(1, uintptr(unsafe.Pointer(&device)))
	//获取下标0的的显卡温度
	var temp uint
	tempProc := nvml.NewProc("nvmlDeviceGetTemperature")
	_, _, _ = tempProc.Call(uintptr(device), uintptr(TemperatureGPU), uintptr(unsafe.Pointer(&temp)))
	return temp
}

// 调速
func FanControl() {
	//验证可以通过上面获取到的显卡温度来进行判断
	// if temp > 40 {
	// 	fmt.Println("red")
	// } else if temp <= 40 {
	// 	fmt.Println("green")
	// } else {
	// 	fmt.Println("0")
	// }

	//测试用
	// cmd1 := exec.Command(`C:\ipmi\64bit\IPMICFG-Win.exe`, "-fan", "1")
	// out, err := cmd1.CombinedOutput()
	// if err != nil {
	// 	fmt.Printf("combined out:\n%s\n", string(out))
	// 	log.Fatalf("cmd.Run() failed with %s\n", err)
	// }
	// fmt.Printf("combined out:\n%s\n", string(out))

	// cmd2 := exec.Command(`C:\ipmi\64bit\IPMICFG-Win.exe`, "-raw", "0x30", "0x70", "0x66", "0x01", "0x01", "0x64")
	// out1, err := cmd2.CombinedOutput()
	// if err != nil {
	// 	fmt.Printf("combined out:\n%s\n", string(out1))
	// 	log.Fatalf("cmd.Run() failed with %s\n", err)
	// }
	// fmt.Printf("combined out:\n%s\n", string(out))

	temp := GetTemp()
	//通过显卡温度调整风扇转速
	if temp > 75 {
		//先将风扇模式调至full，防止下一步调至转速后被IPMI复位，下列地址为IPMICFG的位置，可根据自己服务器的情况来更改
		cmd1 := exec.Command(`C:\ipmi\64bit\IPMICFG-Win.exe`, "-fan", "1")
		err := cmd1.Run()
		if err != nil {
			log.Fatalf("cmd.Run() failed with %s\n", err)
		}
		//风扇转速拉满
		cmd2 := exec.Command(`C:\ipmi\64bit\IPMICFG-Win.exe`, "-raw", "0x30", "0x70", "0x66", "0x01", "0x01", "0x64")
		err2 := cmd2.Run()
		if err != nil {
			log.Fatalf("cmd.Run() failed with %s\n", err2)
		}
		fmt.Printf("\rTEMP:%v FAN:FULL", temp)
	} else if 75 >= temp && temp > 65 {
		//先将风扇模式调至full，防止下一步调至转速后被IPMI复位
		cmd1 := exec.Command(`C:\ipmi\64bit\IPMICFG-Win.exe`, "-fan", "1")
		err := cmd1.Run()
		if err != nil {
			log.Fatalf("cmd.Run() failed with %s\n", err)
		}
		//风扇转速拉至80%
		cmd2 := exec.Command(`C:\ipmi\64bit\IPMICFG-Win.exe`, "-raw", "0x30", "0x70", "0x66", "0x01", "0x01", "0x50")
		err2 := cmd2.Run()
		if err != nil {
			log.Fatalf("cmd.Run() failed with %s\n", err2)
		}
		fmt.Printf("\rTEMP:%v FAN:80	", temp)
	} else if 65 >= temp {
		cmd := exec.Command(`C:\ipmi\64bit\IPMICFG-Win.exe`, "-fan", "2")
		err := cmd.Run()
		if err != nil {
			log.Fatalf("cmd.Run() failed with %s\n", err)
		}
		fmt.Printf("\rTEMP:%v FAN:AUTO", temp)
	}

	// 其中 ipmicfg -raw 0x30 0x70 0x66 0x01 是固定的，表示要调整风扇转速
	// 0x：表示zone n，超微的主板，有两个zone：zone0表示系统（CPU，对应到x10sdv主板，zone0包括FAN1~FAN3）、zone1表示外围（对应到x10sdv主板，zone1包括FAN4）。如果要调整zone0的风扇，此处就填0x00，如果要调整zone1的风扇，此处填0x01
	// 0x：此处0x表示16进制，n表示百分比，1% - 100%转换成16进制，就是(0x01 ~ 0x64)
}

// func PrintWord() {
// 	fmt.Println("1")
// }

func main() {
	location, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		location = time.FixedZone("CST", 8*3600)
	}
	s := gocron.NewScheduler(location)
	s.Every(1).Seconds().Do(func() {
		go FanControl()
		// fmt.Println("Done!")
	})

	s.StartBlocking()
}
