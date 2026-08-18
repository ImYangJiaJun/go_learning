package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// 专门用于读写演示的测试文件
const testFile = "./file_test_demo.txt"

// 专门用于目录操作演示的测试目录
const testDir = "./dir_test_demo"

/*
权限代码说明（八进制，仅 Linux/macOS 生效，Windows 下基本被忽略）：
三位数字从左到右依次对应：文件所有者(owner)、同组用户(group)、其他用户(other)。
每位数字是 读r(4) + 写w(2) + 执行x(1) 的组合，例如 7 = 4+2+1 = rwx。

常用代码：
	0777 = rwxrwxrwx  所有人可读写执行（最宽松，一般不要用）
	0755 = rwxr-xr-x  所有者可读写执行，其他人只读和执行（目录常用）
	0700 = rwx------  仅所有者可读写执行（私有目录常用）
	0666 = rw-rw-rw-  所有人可读写（文件默认权限）
	0644 = rw-r--r--  所有者可读写，其他人只读（文件最常用）
	0600 = rw-------  仅所有者可读写（密钥等敏感文件常用）
对应符号常量：os.ModePerm = 0777；也可用 os.Chmod 修改已有文件的权限。
*/

func main() {
	// 方式一：os.Create + Write 写入 / os.Open + Read 循环读取（自己控制字节数，适合大文件）
	writeBySlice()
	readBySlice()

	// 方式二：bufio 按行写入 / 按行读取（适合逐行处理文本文件）
	writeByBufio()
	readByBufio()

	// 方式三：os.WriteFile 一次性写入 / os.ReadFile 一次性读取（最简单，适合小文件）
	writeByWriteFile()
	readByReadFile()

	// 复制 / 删除文件
	copyFile()
	deleteFile()

	// 目录操作：创建、遍历、重命名、删除
	dirOperations()
}

// 复制文件：打开源文件、创建目标文件，用 io.Copy 流式拷贝（适合大文件）
func copyFile() {
	// 打开源文件
	src, err := os.Open(testFile)
	if err != nil {
		fmt.Println("打开源文件失败：", err)
		return
	}
	defer closeFile(src)

	// 创建目标文件
	dst, err := os.Create("./file_test_copy.txt")
	if err != nil {
		fmt.Println("创建目标文件失败：", err)
		return
	}
	defer closeFile(dst)

	// io.Copy 内部带缓冲，自动分块拷贝，不需要自己读写
	if _, err := io.Copy(dst, src); err != nil {
		fmt.Println("复制文件失败：", err)
		return
	}
	fmt.Println("复制文件完成")
}

// 删除文件：os.Remove 删除指定文件
func deleteFile() {
	if err := os.Remove("./file_test_copy.txt"); err != nil {
		fmt.Println("删除文件失败：", err)
		return
	}
	fmt.Println("删除文件完成")
}

// closeFile 关闭文件并统一处理关闭错误
// 尤其是写入场景，Close 时才会真正刷盘，错误不能忽略
func closeFile(file *os.File) {
	if err := file.Close(); err != nil {
		fmt.Println("文件流关闭失败：", err)
	}
}

// 方式一：os.Open + Read 循环读取
// 自己控制缓冲区大小，内存占用固定，适合读取大文件
func readBySlice() {
	// 只读方式打开当前目录下的文件
	file, err := os.Open(testFile)
	if err != nil {
		fmt.Println("打开文件失败：", err)
		return
	}
	// 确保文件最终会被关闭，并处理关闭错误
	defer closeFile(file)

	var tempSlice = make([]byte, 128) // 每次读取128字节
	for {
		n, err := file.Read(tempSlice)
		if err == io.EOF {
			fmt.Println("读取完毕")
			break
		}
		if err != nil {
			fmt.Println("读取失败：", err)
			return
		}
		fmt.Println("读取到了", n, "字节")
		fmt.Println(string(tempSlice[:n])) // 只输出本次实际读取的内容
	}
}

// 方式二：bufio 按行读取
// 借助 bufio.Reader 的缓冲，可以方便地按行读取文本文件
func readByBufio() {
	file, err := os.Open(testFile)
	if err != nil {
		fmt.Println("打开文件失败：", err)
		return
	}
	defer closeFile(file)

	reader := bufio.NewReader(file)
	for {
		// 以换行符为界读取一行，'\n' 也会被包含在结果中
		line, err := reader.ReadString('\n')
		if err == io.EOF {
			// 文件末尾可能没有换行符，需要把最后一行也输出
			if len(line) > 0 {
				fmt.Print(line)
			}
			fmt.Println("读取完毕")
			break
		}
		if err != nil {
			fmt.Println("读取失败：", err)
			return
		}
		fmt.Print(line)
	}
}

// 方式三：os.ReadFile 一次性读取
// 一行代码读出全部内容，简单方便，但大文件会占用较多内存
func readByReadFile() {
	content, err := os.ReadFile(testFile)
	if err != nil {
		fmt.Println("读取文件失败：", err)
		return
	}
	fmt.Println(string(content))
}

// 方式一：os.Create + Write 循环写入
// 自己控制每次写入的内容，适合大数据量的分块写入
func writeBySlice() {
	// 创建文件，已存在则清空；os.Create 默认权限为 0666
	file, err := os.Create(testFile)
	if err != nil {
		fmt.Println("创建文件失败：", err)
		return
	}
	defer closeFile(file)

	// 分多次写入内容
	lines := []string{"第一行：os.Create + Write 写入\n", "第二行：自己控制每次写入的内容\n"}
	for _, line := range lines {
		_, err := file.WriteString(line)
		if err != nil {
			fmt.Println("写入失败：", err)
			return
		}
	}
}

// 方式二：bufio 缓冲写入
// 先写入内存缓冲区，最后 Flush 一次性落盘，减少磁盘 IO 次数
func writeByBufio() {
	file, err := os.Create(testFile)
	if err != nil {
		fmt.Println("创建文件失败：", err)
		return
	}
	defer closeFile(file)

	writer := bufio.NewWriter(file)
	lines := []string{"第一行：bufio 缓冲写入\n", "第二行：Flush 之后才会真正写入磁盘\n"}
	for _, line := range lines {
		_, err := writer.WriteString(line)
		if err != nil {
			fmt.Println("写入失败：", err)
			return
		}
	}
	// 缓冲区中的数据必须手动 Flush 才会写入文件
	if err := writer.Flush(); err != nil {
		fmt.Println("Flush 失败：", err)
	}
}

// 方式三：os.WriteFile 一次性写入
// 一行代码写入全部内容，简单方便，适合小文件
func writeByWriteFile() {
	content := "第一行：os.WriteFile 一次性写入\n第二行：最简单的方式\n"
	// 文件不存在则创建，存在则清空；0666 为文件权限
	if err := os.WriteFile(testFile, []byte(content), 0666); err != nil {
		fmt.Println("写入文件失败：", err)
	}
}

// 目录操作：创建、判断存在、遍历、重命名、删除
func dirOperations() {
	// 创建单级目录，已存在会报错；0755 为目录权限
	if err := os.Mkdir(testDir, 0755); err != nil {
		fmt.Println("创建目录失败：", err)
	}

	// 创建多级目录，路径中间的目录不存在时会一并创建，已存在也不报错
	if err := os.MkdirAll(testDir+"/sub/inner", 0755); err != nil {
		fmt.Println("创建多级目录失败：", err)
	}

	// 判断文件或目录是否存在
	if _, err := os.Stat(testDir); err == nil {
		fmt.Println(testDir, "存在")
	} else if os.IsNotExist(err) {
		fmt.Println(testDir, "不存在")
	}

	// 在目录下放两个文件，便于演示遍历
	_ = os.WriteFile(testDir+"/a.txt", []byte("a"), 0666)
	_ = os.WriteFile(testDir+"/sub/b.txt", []byte("b"), 0666)

	// os.ReadDir 读取目录下一层的所有条目（不递归）
	entries, err := os.ReadDir(testDir)
	if err != nil {
		fmt.Println("读取目录失败：", err)
		return
	}
	for _, entry := range entries {
		// IsDir 区分条目是目录还是文件
		fmt.Println("遍历到：", entry.Name(), "是否目录：", entry.IsDir())
	}

	// filepath.WalkDir 递归遍历目录下的所有文件和子目录
	err = filepath.WalkDir(testDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		fmt.Println("递归遍历到：", path)
		return nil
	})
	if err != nil {
		fmt.Println("递归遍历失败：", err)
	}

	// 重命名（也可用于移动）目录
	if err := os.Rename(testDir+"/a.txt", testDir+"/a_renamed.txt"); err != nil {
		fmt.Println("重命名失败：", err)
	}

	// os.Remove 只能删除空目录或单个文件；删除非空目录需要用 RemoveAll
	if err := os.RemoveAll(testDir); err != nil {
		fmt.Println("删除目录失败：", err)
	}
}
