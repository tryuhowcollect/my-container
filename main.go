package main

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

func main() {
	if len(os.Args) < 3 {
		panic("引数が足りません")
	}
	switch os.Args[1] {
	case "run":
		run()
	case "child":
		child()
	default:
		panic("引数が不正です")
	}
}

// child()関数が新しい名前空間内で実行され、ホストシステムと隔離された環境で/bin/bash シェルを実行できる
func run() {
	fmt.Printf("Running %v\n", os.Args[2:])
	// exec.Commandで新しいプロセスを作成(/simple_container child /bin/bash)
	cmd := exec.Command("/proc/self/exe", append([]string{"child"}, os.Args[2:]...)...)
	// 新しいプロセスの標準入力、標準出力、標準エラー出力を現在のプロセスに接続
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	// 3つのnamespaceを作成(Mount, UTS, PID)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID,
	}
	// 3つのnamespaceの中で新しいプロセスを実行
	if err := cmd.Run(); err != nil {
		fmt.Printf("Error running command: %v\n", err)
	}
}

func child() {
	fmt.Printf("Running %v\n", os.Args[2:])
	// NEWPIDに対応
	cmd := exec.Command(os.Args[2], os.Args[3:]...) // /bin/bash
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	// 新しいUTS名前空間内でホスト名を "container" に設定(独自のホスト名)
	syscall.Sethostname([]byte("container"))

	setupNetwork()

	// 新しいルートディレクトリを設定
	//syscall.Chroot("/")
	// 現在の作業ディレクトリを新しいルートに変更(上記で指定したディレクトリからの相対パス)
	//os.Chdir("/")
	// proc ファイルシステムを /proc にマウント
	syscall.Mount("proc", "/proc", "proc", 0, "")

	if err := cmd.Run(); err != nil {
		fmt.Printf("Error running command: %v\n", err)
	}
	// プロセス終了後に proc ファイルシステムをアンマウント
	syscall.Unmount("/proc", 0)
}

func setupNetwork() {
	// veth pair の作成
	if err := exec.Command("ip", "link", "add", "veth0", "type", "veth", "peer", "name", "veth1").Run(); err != nil {
		fmt.Printf("Error creating veth pair: %v\n", err)
	}

	// veth1 をネットワーク名前空間に所属させる
	if err := exec.Command("ip", "link", "set", "veth1", "netns", fmt.Sprintf("%d", os.Getpid())).Run(); err != nil {
		fmt.Printf("Error setting veth1 to namespace: %v\n", err)
	}

	// veth0 に IP アドレスを割り当てて UP する
	if err := exec.Command("ip", "addr", "add", "192.168.1.1/24", "dev", "veth0").Run(); err != nil {
		fmt.Printf("Error assigning IP to veth0: %v\n", err)
	}
	if err := exec.Command("ip", "link", "set", "veth0", "up").Run(); err != nil {
		fmt.Printf("Error setting veth0 up: %v\n", err)
	}

	// 名前空間内で veth1 に IP アドレスを割り当てて UP する
	if err := exec.Command("ip", "addr", "add", "192.168.1.2/24", "dev", "veth1").Run(); err != nil {
		fmt.Printf("Error assigning IP to veth1: %v\n", err)
	}
	if err := exec.Command("ip", "link", "set", "veth1", "up").Run(); err != nil {
		fmt.Printf("Error setting veth1 up: %v\n", err)
	}

	// veth pair 間で ping を確認
	if err := exec.Command("ping", "-c", "4", "192.168.1.2").Run(); err != nil {
		fmt.Printf("Error pinging between veth pair: %v\n", err)
	}
}
