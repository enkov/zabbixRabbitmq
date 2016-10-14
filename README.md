ZabbixRabbitMQ
=======================

Send rabbitmq stats to zabbix trapper

Can send info about rabbitmq server, queue, vhosts 

###Build
```
git clone https://github.com/enkov/zabbixRabbitmq.git
go get
go build
```

###Usage
First of all you should change setting in config.
 
Set rabbitmq and zabbix credentials.

Optionally you can use hashicorp vault to get(set vault_enabled in conf) 
- rabbitmq_host(in vault uses name host)
- rabbitmq_user(in vault uses name user)
- rabbitmq_passwd(in vault uses name passwd)

####For example:
If vault path is secret/rabbitmq you should write credentials in it as shown below
```
vault write secret/rabbitmq passwd='secret' user='rabbit' host='192.168.1.77'
```

The second thing to do is to import template in zabbix.
This template realizes only rabbitmq server stats.
If you want to monitor queue or vhosts, when you should add them in zabbix yourself.

####Queue keys fo zabbix trapper
- rabbitmq.queue.+Name(type: int) - show count of all messages in queue
- rabbitmq.queue.+Name.vhost(type: string)
- rabbitmq.queue.+Name.durable(type: bool)
- rabbitmq.queue.+Name.autodelete(type: bool)
- rabbitmq.queue.+Name.status(type: string)
- rabbitmq.queue.+Name.memory(type: int)
- rabbitmq.queue.+Name.consumers(type: int)
- rabbitmq.queue.+Name.policy(type: string)
- rabbitmq.queue.+Name.messagesdetails(type: float)
- rabbitmq.queue.+Name.messagesready(type: int)
- rabbitmq.queue.+Name.messagesreadydetails(type: float)
- rabbitmq.queue.+Name.MessagesUnacknowledged(type: int)
- rabbitmq.queue.+Name.MessagesUnacknowledgedDetails(type: float)

where Name - name of queue in zabbix.

####Vhosts keys fo zabbix trapper
- rabbitmq.vhost."+v.Name+".messages(type: int)
- rabbitmq.vhost."+v.Name+".messagesdetails(type: float)
- rabbitmq.vhost."+v.Name+".messagesready(type: int)
- rabbitmq.vhost."+v.Name+".messagesreadydetails(type: float)
- rabbitmq.vhost."+v.Name+".messagesunacknowledged(type: int)
- rabbitmq.vhost."+v.Name+".messagesunacknowledgeddetails(type: float)

where Name - name of vhost in zabbix.(If Vhost name / in zabbix it will be root)

After all settings done to start using just star binary.
By default zabbixRabbitmq will look for config in current folder with name conf.yml
To specify path to config file use `-config` option.

#### If everything works

If zabbixRabbitMQ works you can add it in cron
```
crontab -e
*/1 * * * * /etc/zabbix/scripts/rabbitmq/zabbixRabbitmq -config /etc/zabbix/scripts/rabbitmq/conf.yml
```
