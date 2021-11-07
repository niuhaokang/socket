package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
)

const (
	addr = "121.41.112.176:9999"
)

func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// cat 接受的数据流是%d|%s格式分别代表数据字节数与数据
func cat(conn net.Conn, filename string) {
	fmt.Fprintf(conn, "cat|1|%s|", filename)

	reader := bufio.NewReader(conn)
	sizeText, err := reader.ReadString('|')
	if err != nil {
		log.Printf("cat读取错误'|': %#v\n", err)
	}
	size, err := strconv.Atoi(sizeText[:len(sizeText)-1])
	if err != nil {
		log.Printf("cat的size转换错误: %#v\n", err)
	}

	if size > 0 {
		ctx := make([]byte, 1024 * 1024)
		n, err := reader.Read(ctx)
		if err != nil {
			log.Fatalf("cat读取异常: %#v\n", err)
		}
		fmt.Println(string(ctx[:n]))
	} else if size == 0 {
		fmt.Println("文件不存在")
	} else {
		fmt.Println("文件不确定是否存在")
	}
}

func ls(conn net.Conn) {
	fmt.Fprint(conn, "ls|0|")

	reader := bufio.NewReader(conn)
	sizeText, err := reader.ReadString('|')
	if err != nil {
		log.Printf("ls解析sizeText异常: %#v\n", err)
	}
	size, err := strconv.Atoi(sizeText[:len(sizeText) - 1])
	if err != nil {
		log.Printf("ls解析size异常: %#v\n", err)
	}
	for size > 0 {
		name, err := reader.ReadString(':')
		if err != nil {
			log.Printf("ls解析文件名错误: %#v\n", err)
		}
		fmt.Println(name[:len(name)-1])
		size--
	}
}

func scp(conn net.Conn, remote string, local string) {
	f, err := PathExists(local)

	if f && err == nil {
		file, err := os.Open(local)
		if err != nil {
			log.Printf("scp打开文件异常: %#v\n", err)
		}
		defer file.Close()

		fmt.Fprintf(conn, "scp|1|%s|", remote)

		ctx := make([]byte, 1024 * 1024)
		n, _ := file.Read(ctx)

		fmt.Fprint(conn, string(ctx[:n]))

		// 表示是否创建成功
		reader := bufio.NewReader(conn)
		flag, err := reader.ReadString('|')
		if err != nil {
			log.Printf("scp创建文件异常: %#v\n", err)
		}
		if flag[:len(flag) -1] == "0" {
			fmt.Println("Uploaded successfully")
		} else if flag[:len(flag) -1] == "1"{
			fmt.Println("文件已存在")
		} else {
			fmt.Println("Upload failed")
		}
	} else {
		fmt.Println("上传文件不存在")
	}
}

func wget(conn net.Conn, remote string, local string) {
	flag, err := PathExists(local)
	reader := bufio.NewReader(conn)
	if flag && err == nil {
		// 本地已存在该文件
		fmt.Println("文件已存在!")
	} else {
		fmt.Fprintf(conn, "wget|1|%s|", remote)

		f, err := reader.ReadString('|')
		if err != nil {
			log.Printf("wget文件解析错误")
		}
		if f[:len(f)-1] == "0" {
			reader := bufio.NewReader(conn)
			file, err := os.Create(local)
			if err != nil {
				log.Printf("scp创建文件异常: %#v\n", err)
			}
			defer file.Close()

			ctx := make([]byte, 1024 * 1024)
			n, _ := reader.Read(ctx)
			file.Write(ctx[:n])
		} else {
			fmt.Println("文件不存在！")
		}

	}
}

func main() {
	// 定义日志文件
	logfile, err := os.OpenFile("client.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, os.ModePerm)
	if err != nil {
		log.Fatalf("创建或打开日志文件失败: %#v\n", err)
	}
	defer logfile.Close()
	log.SetOutput(logfile)

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		log.Fatalf("连接服务器异常: %#v\n", err)
	}
	defer conn.Close()

	scanner := bufio.NewScanner(os.Stdin)
	END:
		for {
			fmt.Print("请输入指令:")
			scanner.Scan()
			input := scanner.Text()
			cmds := strings.Split(input, " ")
			switch cmds[0] {
			case "cat":
				cat(conn, cmds[1])
			case "ls":
				ls(conn)
			case "scp":
				if len(cmds) != 3 {
					fmt.Println("指令格式错误！")
					continue
				}
				scp(conn, cmds[1], cmds[2])
			case "wget":
				if len(cmds) != 3 {
					fmt.Println("指令格式错误！")
					continue
				}
				wget(conn, cmds[1], cmds[2])
			case "exit":
				fmt.Fprintf(conn, "exit|0|")
				break END
			default :
				fmt.Println("指令无法识别")
			}
		}
}
