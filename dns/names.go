package dns

import (
	"bufio"
	"net"
	"os"
	"strings"
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/micro-plat/ddns/dns/query"
	"github.com/micro-plat/lib4go/file"
	"github.com/micro-plat/lib4go/logger"
)

var defNames = []string{"114.114.114.114", "8.8.8.8"}

//Names 本地name server读取配置
type Names struct {
	closeCh chan struct{}
	watcher *fsnotify.Watcher
	log     logger.ILogger
	names   []string
	lk      sync.Mutex
}

//NewNames 创建本地host文件读取对象
func NewNames(log logger.ILogger) *Names {
	names := &Names{
		closeCh: make(chan struct{}),
		log:     log,
	}
	return names
}

//Start 启动服务，进行本地文件读取与文件变动重新加载
func (f *Names) Start() (err error) {
	if err := f.checkAndCreateConf(); err != nil {
		return err
	}
	f.watcher, err = fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	if err := f.watcher.Add(query.NAME_ROOT); err != nil {
		return err
	}
	err = f.reload()
	if err != nil {
		return err
	}
	f.log.Infof("[启用 NAMES,%d条]", f.len())
	go f.loopWatch()
	return nil
}

//Lookup 查询域名解析结果
func (f *Names) Lookup() []string {
	f.lk.Lock()
	defer f.lk.Unlock()
	return f.names
}
func (f *Names) len() int {
	f.lk.Lock()
	defer f.lk.Unlock()
	return len(f.names)
}

//Close 关闭服务
func (f *Names) Close() error {
	close(f.closeCh)
	if f.watcher != nil {
		f.watcher.Close()
	}
	return nil
}

func (f *Names) loopWatch() {
	for {
		select {
		case <-f.closeCh:
			return
		case event := <-f.watcher.Events:
			if event.Name != query.NAME_FILE {
				continue
			}
			switch event.Op {
			case fsnotify.Write:
				f.log.Infof("文件%s发生变更", event.Name)
				if err := f.reload(); err != nil {
					f.log.Error(err)
				}
				f.log.Infof("[启用 NAMES,%d]", f.len())
			default:

			}
		}
	}
	f.log.Info("exit")
}

func (f *Names) reload() error {
	names, err := f.load(query.NAME_FILE)
	if err != nil {
		return err
	}
	nNames := f.sortByTTL(names)
	for i, nm := range nNames {
		nNames[i] = net.JoinHostPort(nm, "53")
	}
	f.lk.Lock()
	defer f.lk.Unlock()
	f.names = nNames
	return nil
}
func (f *Names) sortByTTL(names []string) []string {
	sorted, err := getSortedServer(names...)
	if err != nil {
		return names
	}
	return sorted
}

//load 加载配置文件，并读取指定的文件内容
func (f *Names) load(path string) ([]string, error) {
	buf, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer buf.Close()

	ips := make(map[string]string)
	names := make([]string, 0, 1)
	scanner := bufio.NewScanner(buf)
	for scanner.Scan() {

		line := strings.Replace(strings.TrimSpace(scanner.Text()), "\t", " ", -1)
		if strings.HasPrefix(line, "#") || line == "" {
			continue
		}

		ip := net.ParseIP(line)
		if ip == nil {
			continue //ip格式错误
		}
		if _, ok := ips[line]; !ok {
			ips[line] = line
			names = append(names, line)
		}

	}
	return names, nil
}
func (f *Names) checkAndCreateConf() error {
	_, err := os.Stat(query.NAME_FILE)
	if err == nil {
		return nil
	}
	if !os.IsNotExist(err) {
		return err
	}
	fwriter, err := file.CreateFile(query.NAME_FILE)
	if err != nil {
		return err
	}

	defer fwriter.Close()
	_, err = fwriter.Write([]byte(strings.Join(defNames, "\n")))
	if err != nil {
		return err
	}
	return nil
}
