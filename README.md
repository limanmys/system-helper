# Çalıştırma

Herhangi bir bağımlığa gerek olmadan direkt çalıştırabilirsiniz. Servisi çalıştırmak için : 

`sudo build/liman_system`

#### Test Edilen İşletim Sistemleri : 

Debian 10
Pardus 19
CentOS 8 Minimal Server
Fedora Workstation 32

#### Sorgu Örnekleri : 

##### Kullanıcı Ekleme

extensionId verisini düzenleyerek : 

`curl '127.0.0.1:1803/userAdd?extensionId=mert'`

##### Kullanıcı Silme

extensionId verisini düzenleyerek : 

`curl '127.0.0.1:1803/userRemove?extensionId=mert'`

##### DNS Düzenleme
server1,server2 ve server3 verilerini düzenleyerek : 

`curl '127.0.0.1:1803/dns?server1=192.168.0.1&server2=8.8.8.8&server3=8.8.4.4'`

##### Sertifika Ekleme
tmpPath,targetName verilerini düzenleyerek : 
Not: tmpPath parametresi encoded olmalı, encode etmek için : https://www.urlencoder.org/

`curl '127.0.0.1:1803/certificateAdd?tmpPath=%2Ftmp%2Ftest.crt&targetName=samba_dc_636'`

##### Sertifika Silme
targetName verisini düzenleyerek : 
`curl '127.0.0.1:1803/certificateRemove?targetName=samba_dc_636'`

##### Liman Eklenti İzinlerini Düzenlemek

extensionId ve extensionName verilerini düzenleyerek : 

`curl '127.0.0.1:1803/fixPermissions?extensionId=c9329ce0-c23c-11ea-b3de-0242ac130004&extensionName=mert'`