from pyzabbix import ZabbixAPI

CREATE_QUEUE_GROUP = True
CREATE_VHOSTS_GROUP = True

TEMPALTE_NAME = 'Template App RabbitMQ'

ZABBIX_USER = 'user'
ZABBIX_PASSWD = 'passwd'
TEMPALTE_ID = ''
QUEUE_APPLICATION_ID = []
VHOST_APPLICATION_ID = []

QUEUES = ['test']
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

if TEMPALTE_ID == '':
    # I don't know if Templates groupid allways 1 so I get it from zabbix
    hostgroups = zapi.hostgroup.get()
    group_id = 0
    for i in hostgroups:
        if i['name'] == 'Templates':
            group_id = i['groupid']
            break
    template_id = zapi.template.create(
                      groups=group_id,
                      host=TEMPALTE_NAME,
                      name=TEMPALTE_NAME)
    TEMPALTE_ID = template_id['templateids'][0]

if CREATE_QUEUE_GROUP:
    queue_appid = zapi.application.create(
        hostid=TEMPALTE_ID,
        name='Queues')
    QUEUE_APPLICATION_ID.append(queue_appid['applicationids'][0])

if CREATE_VHOSTS_GROUP:
    vhost_appid = zapi.application.create(
        hostid=TEMPALTE_ID,
        name='Vhosts')
    VHOST_APPLICATION_ID.append(vhost_appid['applicationids'][0])

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
