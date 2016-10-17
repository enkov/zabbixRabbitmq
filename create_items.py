from pyzabbix import ZabbixAPI

ZABBIX_USER = 'user'
ZABBIX_PASSWD = 'passwd'

TEMPALTE_ID = "10264"
QUEUE_APPLICATION_ID = ['123']
VHOST_APPLICATION_ID = ['123']

QUEUES = ['']
QUEUE_KEYS = {'messages': 3, 'vhost': 4, 'durable': 4, 'autodelete': 4,
              'status': 4, 'memory': 3, 'consumers': 3, 'policy': 4,
              'messagesdetails': 0, 'messagesready': 3, 'messagesreadydetails': 0,
              'messagesunacknowledged': 3, 'messagesunacknowledgeddetails': 0
              }

VHOSTS = ['root']
VHOSTS_KEYS = {'messages': 3, 'messagesdetails': 0, 'messagesready': 3,
               'messagesreadydetails': 0, 'messagesunacknowledged': 3,
               'messagesunacknowledgeddetails': 0
               }

zapi = ZabbixAPI("http://example.com/zabbix")
zapi.login(ZABBIX_USER, ZABBIX_PASSWD)

# Create items for rabbitmq queues
for i in QUEUES:
    for k, v in QUEUE_KEYS.iteritems():
        zapi.item.create(
            hostid=TEMPALTE_ID,
            name='RabbitMQ Queue ' + i.upper() + ' ' + k,
            key_='rabbitmq.queue.' + i + '.' + k,
            type=2,
            value_type=v,
            applications=QUEUE_APPLICATION_ID
        )

# Create items for rabbitmq vhosts
for i in VHOSTS:
    for k, v in VHOSTS_KEYS.iteritems():
        zapi.item.create(
            hostid=TEMPALTE_ID,
            name='RabbitMQ Vhost ' + i.upper() + ' ' + k,
            key_='rabbitmq.vhost.' + i + '.' + k,
            type=2,
            value_type=v,
            applications=VHOST_APPLICATION_ID
        )
