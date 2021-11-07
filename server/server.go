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
	addr = ":9999"
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

// getCmd 用于解析服务端传输的指令并解码为指令和字符串切片
func getCmd(conn net.Conn) (string, []string) {
	reader := bufio.NewReader(conn)

	op, err := reader.ReadString('|')
	if err != nil {
		log.Printf("解析cmd操作符异常: %#v\n", err)
	}

	cntText, err := reader.ReadString('|')

	if err != nil {
		log.Printf("解析cmd参数数量字符串异常: %#v\n", err)
	}

	cnt, err := strconv.Atoi(cntText[:len(cntText) - 1])
	if err != nil {
		log.Printf("解析cmd参数数量异常: %#v\n", err)
	}

	args := make([]string, 0, cnt)

	for cnt > 0 {
		params, err := reader.ReadString('|')
		if err != nil {
			log.Printf("解析cmd参数异常: %#v\n", err)
		}
		args = append(args, params[:len(params) - 1])
		cnt--
	}
	return op[:len(op) - 1], args
}

func cat(conn net.Conn, filename string) {
	fileExist, err := PathExists(filename)
	// 文件存在
	if fileExist && err == nil{
		file, err := os.Open(filename)
		defer file.Close()
		if err != nil {
			log.Printf("打开文件错误: %#v\n", err)
		}
		ctx := make([]byte, 1024 * 1024)
		n, _ := file.Read(ctx)
		fmt.Fprintf(conn, "%d|%s", n, string(ctx[:n]))
	} else if !fileExist && err == nil {
		// 文件不存在
		fmt.Fprint(conn, "0||")
	} else {
		fmt.Fprint(conn, "-1||")
	}
}

func ls(conn net.Conn) {
	file, _ := os.Open(".")
	defer file.Close()

	// 返回file指向文件夹下的所有文件名
	names, _ := file.Readdirnames(-1)

	suffix := ""
	if len(names) > 0 {
		suffix = ":"
	}

	fmt.Fprintf(conn, "%d|%s%s", len(names), strings.Join(names, ":"), suffix)
}

func scp(conn net.Conn, filename string) {
	flag, err := PathExists(filename)
	if flag && err == nil {
		fmt.Fprint(conn, "1|")
	} else {
		reader := bufio.NewReader(conn)
		file, err := os.Create(filename)
		if err != nil {
			log.Printf("scp创建文件异常: %#v\n", err)
		}
		defer file.Close()

		ctx := make([]byte, 1024 * 1024)
		n, _ := reader.Read(ctx)
		file.Write(ctx[:n])

		fmt.Fprint(conn, "0|")
	}
}

func wget(conn net.Conn, filename string) {
	flag, err := PathExists(filename)
	fmt.Println(filename)
	if flag && err == nil {
		fmt.Fprint(conn, "0|")
		file, err := os.Open(filename)
		if err != nil {
			log.Printf("wget打开文件错误: %#v\n", err)
		}
		defer file.Close()

		ctx := make([]byte, 1024 * 1024)
		n, _ := file.Read(ctx)

		fmt.Fprint(conn, string(ctx[:n]))
	} else {
		fmt.Fprint(conn, "1|")
	}
}

// handleConn 处理连接函数
func handleConn(conn net.Conn) {
	defer conn.Close()
	END:
		for {
			op, args := getCmd(conn)
			switch op {
			case "cat":
				cat(conn, args[0])
			case "ls":
				ls(conn)
			case "scp":
				scp(conn, args[0])
			case "wget":
				wget(conn, args[0])
			case "exit":
				break END
			}
		}
}

func main() {
	// 定义日志文件
	logfile, err := os.OpenFile("server.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, os.ModePerm)
	if err != nil {
		log.Fatalf("创建或打开日志文件失败: %#v\n", err)
	}
	defer logfile.Close()
	log.SetOutput(logfile)

	// 使用tcp连接监听addr端口
	listener, err := net.Listen("tcp", addr)
	if err != nil {

	}
	defer listener.Close()

	log.Printf("已监听端口: %s\n", addr)
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("连接异常: %#v\n", err)
			continue
		}
		log.Printf("已获取连接: %s\n", conn.RemoteAddr())
		go handleConn(conn)
	}
}
