package main

import (
	"flag"
	"strconv"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	zabbix "github.com/blacked/go-zabbix"
	"github.com/michaelklishin/rabbit-hole"
	"github.com/spf13/viper"
)

var (
	wg sync.WaitGroup
)

func main() {
	var vaultCfg map[string]interface{}
	var confFile, rabbitPasswd, rabbitUser, rabbitHost, rabbitPort string
	flag.StringVar(&rabbitPasswd, "passwd", "", "rabbitmq passwd")
	flag.StringVar(&rabbitUser, "username", "", "rabbitmq user")
	flag.StringVar(&rabbitHost, "host", "", "rabbitmq host")
	flag.StringVar(&rabbitPort, "port", "", "rabbitmq api port")
	flag.StringVar(&confFile, "config", "conf.yml", "config file path")
	flag.Parse()

	viper.SetConfigFile(confFile)

	err := viper.ReadInConfig()
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Fatal("No configuration file loaded - using defaults")
	}

	// Configure logging
	logLevel, err := log.ParseLevel(viper.GetString("log_level"))
	customFormatter := new(log.TextFormatter)
	log.SetFormatter(customFormatter)
	customFormatter.FullTimestamp = true
	log.SetLevel(logLevel)

	// Read config
	vaultRabbitPath := viper.GetString("vault_rabbit_path")
	vaultAddr := viper.GetString("vault_addr")
	vaultToken := viper.GetString("vault_token")
	vaultEnabled := viper.GetBool("vault_enabled")
	zabbixAgentHostname := viper.GetString("zabbix_agent_hostname")
	zabbixHost := viper.GetString("zabbix_host")
	zabbixPort := viper.GetInt("zabbix_port")

	rabbitmqNodeName := viper.GetString("rabbitmq_node_name")
	viper.SetDefault("rabbitmq_api_port", "15672")
	rabbitPort = viper.GetString("rabbitmq_api_port")

	// Read secrets from vault
	if vaultEnabled {
		log.Info("Read rabbitmq conf from vault")
		vaultCfg, err = readCfgVault(vaultRabbitPath, vaultAddr, vaultToken)
		if err != nil {
			log.WithFields(log.Fields{"error": err}).Warn("Can't read from vault")
		}
		rabbitPasswd = vaultCfg["passwd"].(string)
		rabbitUser = vaultCfg["user"].(string)
		rabbitHost = vaultCfg["host"].(string)
	} else {
		log.Info("Read rabbitmq conf from config file")
		rabbitPasswd = viper.GetString("rabbitmq_passwd")
		rabbitUser = viper.GetString("rabbitmq_user")
		rabbitHost = viper.GetString("rabbitmq_host")
	}

	// Create rabbitmq client
	log.Debug("Connect to rabbitmq")
	rmqc, err := rabbithole.NewClient("http://"+rabbitHost+":"+rabbitPort, rabbitUser, rabbitPasswd)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Fatalln("Can't connect to rabbit")
	}

	log.Debug("Start sendRabbitOverview goroutine")
	wg.Add(1)
	go sendRabbitOverview(zabbixAgentHostname, rmqc, zabbixHost, zabbixPort)

	log.Debug("Start sendQueueInfo goroutine")
	wg.Add(1)
	go sendQueueInfo(zabbixAgentHostname, rmqc, zabbixHost, zabbixPort)

	log.Debug("Start sendRabbitNodeInfo goroutine")
	wg.Add(1)
	go sendRabbitNodeInfo(rabbitmqNodeName, zabbixAgentHostname, rmqc, zabbixHost, zabbixPort)

	log.Debug("Start sendVhostInfo goroutine")
	wg.Add(1)
	go sendVhostInfo(zabbixAgentHostname, rmqc, zabbixHost, zabbixPort)

	wg.Wait()
}

func sendRabbitNodeInfo(rabbitNode string, hostname string, rmqc *rabbithole.Client, zabbixHost string, zabbixPort int) error {
	defer wg.Done()
	var metrics []*zabbix.Metric
	log.Info("Get info about rabbitmq node")
	node, err := rmqc.GetNode(rabbitNode)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Fatalln("Can't get node info")
	}

	metrics = append(metrics, zabbix.NewMetric(hostname, "rabbitmq.nodeinfo.name", node.Name, time.Now().Unix()))
	metrics = append(metrics, zabbix.NewMetric(hostname, "rabbitmq.nodeinfo.isrunning", strconv.FormatBool(node.IsRunning), time.Now().Unix()))
	metrics = append(metrics, zabbix.NewMetric(hostname, "rabbitmq.nodeinfo.ospid", string(node.OsPid), time.Now().Unix()))
	metrics = append(metrics, zabbix.NewMetric(hostname, "rabbitmq.nodeinfo.fdused", strconv.Itoa(node.FdUsed), time.Now().Unix()))
	metrics = append(metrics, zabbix.NewMetric(hostname, "rabbitmq.nodeinfo.fdtotal", strconv.Itoa(node.FdTotal), time.Now().Unix()))
	metrics = append(metrics, zabbix.NewMetric(hostname, "rabbitmq.nodeinfo.socketsused", strconv.Itoa(node.SocketsUsed), time.Now().Unix()))
	metrics = append(metrics, zabbix.NewMetric(hostname, "rabbitmq.nodeinfo.socketstotal", strconv.Itoa(node.SocketsTotal), time.Now().Unix()))
	metrics = append(metrics, zabbix.NewMetric(hostname, "rabbitmq.nodeinfo.memused", strconv.Itoa(node.MemUsed), time.Now().Unix()))
	metrics = append(metrics, zabbix.NewMetric(hostname, "rabbitmq.nodeinfo.memlimit", strconv.Itoa(node.MemLimit), time.Now().Unix()))
	metrics = append(metrics, zabbix.NewMetric(hostname, "rabbitmq.nodeinfo.memalarm", strconv.FormatBool(node.MemAlarm), time.Now().Unix()))
	metrics = append(metrics, zabbix.NewMetric(hostname, "rabbitmq.nodeinfo.diskfree", strconv.Itoa(node.DiskFree), time.Now().Unix()))
	metrics = append(metrics, zabbix.NewMetric(hostname, "rabbitmq.nodeinfo.diskfreelimit", strconv.Itoa(node.DiskFreeLimit), time.Now().Unix()))
	metrics = append(metrics, zabbix.NewMetric(hostname, "rabbitmq.nodeinfo.diskfreealarm", strconv.FormatBool(node.DiskFreeAlarm), time.Now().Unix()))

	sendToZabbix(zabbixHost, zabbixPort, metrics)
	return err
}

//TODO add MessageStats
func sendQueueInfo(hostname string, rmqc *rabbithole.Client, zabbixHost string, zabbixPort int) error {
	defer wg.Done()
	var metrics []*zabbix.Metric
	log.Info("Get info about rabbitmq queues")
	qs, err := rmqc.ListQueues()
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Fatalln("Can't get list of queues")
	}
	for _, v := range qs {
		metrics = append(metrics, zabbix.NewMetric(hostname, "rabbitmq.queue."+v.Name, strconv.Itoa(v.Messages), time.Now().Unix()))
		metrics = append(metrics, zabbix.NewMetric(hostname, "rabbitmq.queue."+v.Name+".vhost", v.Vhost, time.Now().Unix()))
		metrics = append(metrics, zabbix.NewMetric(hostname, "rabbitmq.queue."+v.Name+".durable", strconv.FormatBool(v.Durable), time.Now().Unix()))
		metrics = append(metrics, zabbix.NewMetric(hostname, "rabbitmq.queue."+v.Name+".autodelete", strconv.FormatBool(v.AutoDelete), time.Now().Unix()))
		metrics = append(metrics, zabbix.NewMetric(hostname, "rabbitmq.queue."+v.Name+".status", v.Status, time.Now().Unix()))
		metrics = append(metrics, zabbix.NewMetric(hostname, "rabbitmq.queue."+v.Name+".memory", strconv.FormatInt(v.Memory, 10), time.Now().Unix()))
		metrics = append(metrics, zabbix.NewMetric(hostname, "rabbitmq.queue."+v.Name+".consumers", strconv.Itoa(v.Consumers), time.Now().Unix()))
		metrics = append(metrics, zabbix.NewMetric(hostname, "rabbitmq.queue."+v.Name+".policy", v.Policy, time.Now().Unix()))
		metrics = append(metrics, zabbix.NewMetric(hostname, "rabbitmq.queue."+v.Name+".messagesdetails", strconv.FormatFloat(float64(v.MessagesDetails.Rate), 'f', -1, 32), time.Now().Unix()))
		metrics = append(metrics, zabbix.NewMetric(hostname, "rabbitmq.queue."+v.Name+".messagesready", strconv.Itoa(v.MessagesReady), time.Now().Unix()))
		metrics = append(metrics, zabbix.NewMetric(hostname, "rabbitmq.queue."+v.Name+".messagesreadydetails", strconv.FormatFloat(float64(v.MessagesReadyDetails.Rate), 'f', -1, 32), time.Now().Unix()))
		metrics = append(metrics, zabbix.NewMetric(hostname, "rabbitmq.queue."+v.Name+".MessagesUnacknowledged", strconv.Itoa(v.MessagesUnacknowledged), time.Now().Unix()))
		metrics = append(metrics, zabbix.NewMetric(hostname, "rabbitmq.queue."+v.Name+".MessagesUnacknowledgedDetails", strconv.FormatFloat(float64(v.MessagesUnacknowledgedDetails.Rate), 'f', -1, 32), time.Now().Unix()))
	}
	log.Debug("Send rabbitmq queues info to zabbix")
	sendToZabbix(zabbixHost, zabbixPort, metrics)
	return err
}

func sendVhostInfo(hostname string, rmqc *rabbithole.Client, zabbixHost string, zabbixPort int) error {
	defer wg.Done()
	var metrics []*zabbix.Metric
	log.Info("Get info about rabbitmq vhosts")
	vh, err := rmqc.ListVhosts()
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Fatalln("Can't get list of queues")
	}
	for _, v := range vh {
		if v.Name == "/" {
			v.Name = "root"
		}
		metrics = append(metrics, zabbix.NewMetric(hostname, "rabbitmq.vhost."+v.Name+".messages", strconv.Itoa(v.Messages), time.Now().Unix()))
		metrics = append(metrics, zabbix.NewMetric(hostname, "rabbitmq.vhost."+v.Name+".messagesdetails", strconv.FormatFloat(float64(v.MessagesDetails.Rate), 'f', -1, 32), time.Now().Unix()))
		metrics = append(metrics, zabbix.NewMetric(hostname, "rabbitmq.vhost."+v.Name+".messagesready", strconv.Itoa(v.MessagesReady), time.Now().Unix()))
		metrics = append(metrics, zabbix.NewMetric(hostname, "rabbitmq.vhost."+v.Name+".messagesreadydetails", strconv.FormatFloat(float64(v.MessagesReadyDetails.Rate), 'f', -1, 32), time.Now().Unix()))
		metrics = append(metrics, zabbix.NewMetric(hostname, "rabbitmq.vhost."+v.Name+".messagesunacknowledged", strconv.Itoa(v.MessagesUnacknowledged), time.Now().Unix()))
		metrics = append(metrics, zabbix.NewMetric(hostname, "rabbitmq.vhost."+v.Name+".messagesunacknowledgeddetails", strconv.FormatFloat(float64(v.MessagesUnacknowledgedDetails.Rate), 'f', -1, 32), time.Now().Unix()))
	}
	log.Debug("Send rabbitmq vhosts info to zabbix")
	sendToZabbix(zabbixHost, zabbixPort, metrics)
	return err
}

func sendRabbitOverview(hostname string, rmqc *rabbithole.Client, zabbixHost string, zabbixPort int) error {
	defer wg.Done()
	var metrics []*zabbix.Metric
	log.Info("Get overview about rabbitmq")
	overview, err := rmqc.Overview()
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Fatalln("Can't get node info")
	}
	metrics = append(metrics, zabbix.NewMetric(hostname, "rabbitmq.info.managementversion", overview.ManagementVersion, time.Now().Unix()))
	//metrics = append(metrics, zabbix.NewMetric(hostname, "rabbitmq.info.statisticslevel", overview.StatisticsLevel, time.Now().Unix()))
	metrics = append(metrics, zabbix.NewMetric(hostname, "rabbitmq.info.rabbitmqversion", overview.RabbitMQVersion, time.Now().Unix()))
	metrics = append(metrics, zabbix.NewMetric(hostname, "rabbitmq.info.erlangversion", overview.ErlangVersion, time.Now().Unix()))
	metrics = append(metrics, zabbix.NewMetric(hostname, "rabbitmq.info.fullerlangversion", overview.FullErlangVersion, time.Now().Unix()))
	metrics = append(metrics, zabbix.NewMetric(hostname, "rabbitmq.info.statisticsdbnode", overview.StatisticsDBNode, time.Now().Unix()))

	log.Debug("Send rabbitmq overview to zabbix")
	sendToZabbix(zabbixHost, zabbixPort, metrics)
	return err
}

func sendToZabbix(zabbixHost string, zabbixPort int, metrics []*zabbix.Metric) {
	packet := zabbix.NewPacket(metrics)

	// Send packet to zabbix
	z := zabbix.NewSender(zabbixHost, zabbixPort)
	z.Send(packet)
}
