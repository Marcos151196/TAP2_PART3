# TASK 4 (FULLY-DISTRIBUTED)

En primer lugar tenemos que levantar 3 instancias EC2 (1 para el master y 2 esclavos). Las crearemos todas en la misma subnet (172.31.32.0/20) y les asignaremos las IPs privadas 
```
172.31.32.10 al master
172.31.32.11 al esclavo 1
172.31.32.12 al esclavo 2
```
Todas las instancias las crearemos con un security group que de acceso TCP All-trafic (para no rallarnos, al fin y al cabo es una práctica), SSH, HTTP y HTTPS. 

A mayores, crearemos una IP elástica (en nuestro caso 52.22.18.120) y se la asignamos al master una vez lanzada la instancia. Esto sólo lo hacemos por comodidad para acceder a las interfaces de monitorización web de hadoop siempre con la misma IP (cunde añadirla a favoritos), pero no es para nada necesario.

NOTA: tanto para el master como los esclavos hemos elegido t2.large (2 cores y 8GB de RAM) para no tener problemas de insuficiencia de recursos y hacer el testing más rápido.

## 1 - Añadir maestro y esclavos a la lista de hosts del maestro y de cada uno de los esclavos:
``` bash
sudo nano /etc/hosts

172.31.32.10 master
172.31.32.11 slave-1
172.31.32.12 slave-2
```
NOTA: si alguno de los archivos contiene "127.0.1.1 host_name", eliminarla.

## 2 - Configurar ssh en el master para dar acceso a master y nodos

En el master (en el ssh-keygen dejar todo vacio):
```bash
ssh-keygen

scp -i /home/ubuntu/tap2.pem /home/ubuntu/.ssh/id_rsa.pub ubuntu@ec2-54-164-129-139.compute-1.amazonaws.com:/home/ubuntu

scp -i /home/ubuntu/tap2.pem /home/ubuntu/.ssh/id_rsa.pub ubuntu@ec2-54-147-110-59.compute-1.amazonaws.com:/home/ubuntu

cat ~/.ssh/id_rsa.pub >> ~/.ssh/authorized_keys
```

En cada uno de los datanodes:

```bash
cat ~/.ssh/id_rsa.pub >> ~/.ssh/authorized_keys
```

En el master:

```bash
sudo nano ~/.ssh/config
```

``` bash
Host master
  HostName 172.31.32.10
  User ubuntu
  IdentityFile /home/ubuntu/tap2.pem

Host slave-1
  HostName ec2-54-164-129-139.compute-1.amazonaws.com
  User ubuntu
  IdentityFile /home/ubuntu/tap2.pem

Host slave-2
  HostName ec2-54-147-110-59.compute-1.amazonaws.com
  User ubuntu
  IdentityFile /home/ubuntu/tap2.pem
```

Comprobación (escribir "yes" y darle enter) \
En el master:
```bash
ssh master
exit
ssh slave-1
exit
ssh slave-2
exit
```

En el master:
```bash
nano /home/ubuntu/hadoop/etc/hadoop/workers
```

```bash
slave-1
slave-2
```

## 3 - Configurar los XMLs
### Master core-site.xml

```xml
<configuration>
	<property>
		<name>fs.defaultFS</name>
		<value>hdfs://master:9000</value>
    </property>
	<property>
		<name>fs.s3a.access.key</name>
		<value>PONER_AQUI_AWS_ACCESS_KEY</value>
	</property>
	<property>
		<name>fs.s3a.secret.key</name>
		<value>PONER_AQUI_AWS_SECRET_KEY</value>
	</property>
	<property>
		<name>fs.AbstractFileSystem.s3a.imp</name>
		<value>org.apache.hadoop.fs.s3a.S3A</value>
	</property>
	<property>
		<name>fs.s3a.endpoint</name>
		<value>s3.us-east-1.amazonaws.com</value>
	</property>
	<property>
		<name>io.compression.codecs</name>
		<value>org.apache.hadoop.io.compress.GzipCodec, org.apache.hadoop.io.compress.DefaultCodec, org.apache.hadoop.io.compress.BZip2Codec, com.hadoop.compression.lzo.LzoCodec, com.hadoop.compression.lzo.LzopCodec</value>
	</property>
	<property>
		<name>io.compression.codec.lzo.class</name>
		<value>com.hadoop.compression.lzo.LzoCodec</value>
	</property>
</configuration>
```

### Master hdfs-site.xml
```xml
<configuration>
	<property>
      <name>dfs.replication</name>
      <value>2</value>
  </property>
  <property>
    <name>dfs.namenode.name.dir</name>
    <value>file:///home/ubuntu/hadoop/hdfs/namenode</value>
  </property>
  <property>
    <name>dfs.datanode.data.dir</name>
    <value>file:///home/ubuntu/hadoop/hdfs/datanode</value>
  </property>
  <property>
    <name>dfs.permissions</name>
    <value>false</value>
  </property>
  <property>
    <name>dfs.datanode.use.datanode.hostname</name>
    <value>false</value>
  </property>
  <property>
    <name>dfs.namenode.datanode.registration.ip-hostname-check</name>
    <value>false</value>
  </property>
</configuration>
```

### Master yarn-site.xml
```xml
<configuration>
	<property>
		<name>yarn.nodemanager.aux-services</name>
		<value>mapreduce_shuffle</value>
  	</property>
  	<property>
		<name>yarn.nodemanager.env-whitelist</name>
		<value>JAVA_HOME,HADOOP_COMMON_HOME,HADOOP_HDFS_HOME,HADOOP_CONF_DIR,CLASSPATH_PREPEND_DISTCACHE,HADOOP_YARN_HOME,HADOOP_MAPRED_HOME</value>
  	</property>
	<property>
		<name>yarn.resourcemanager.hostname</name>
		<value>master</value>
	</property>
</configuration>
```

### Master mapred-site.xml
```xml
<configuration>
	<property>
		<name>mapreduce.framework.name</name>
		<value>yarn</value>
	</property>
	<property>
		<name> mapreduce.jobhistory.address </name>
		<value>master:10020</value>
  	</property>
	<property>
        <name>mapreduce.framework.name</name>
        <value>yarn</value>
    </property>
    <property>
        <name>yarn.app.mapreduce.am.env</name>
        <value>HADOOP_MAPRED_HOME=${HADOOP_HOME}</value>
    </property>
    <property>
        <name>mapreduce.map.env</name>
        <value>HADOOP_MAPRED_HOME=${HADOOP_HOME}</value>
    </property>
    <property>
        <name>mapreduce.reduce.env</name>
        <value>HADOOP_MAPRED_HOME=${HADOOP_HOME}</value>
    </property>
</configuration>
```

### Slaves core-site.xml
```xml
<configuration>
	<property>
		<name>fs.defaultFS</name>
		<value>hdfs://master:9000</value>
	</property>
</configuration>
```

### Slaves hdfs-site.xml
```xml
<configuration>
	<property>
      <name>dfs.replication</name>
      <value>2</value>
  </property>
  <property>
    <name>dfs.datanode.data.dir</name>
    <value>file:///home/ubuntu/hadoop/hdfs/datanode</value>
  </property>
  <property>
    <name>dfs.permissions</name>
    <value>false</value>
  </property>
  <property>
    <name>dfs.datanode.use.datanode.hostname</name>
    <value>false</value>
  </property>
</configuration>
```

### Slaves yarn-site.xml
```xml
<configuration>
  <property>
    <name>yarn.nodemanager.aux-services</name>
    <value>mapreduce_shuffle</value>
  </property>
  <property>
    <name>yarn.resourcemanager.hostname</name>
    <value>master</value>
  </property>
</configuration>
```

### Slaves mapred-site.xml
```xml
<configuration>
  <property>
    <name>mapreduce.framework.name</name>
    <value>yarn</value>
  </property>
</configuration>
```

### ~/.bashrc (igual para master y slaves)
```bash
export PDSH_RCMD_TYPE=ssh

# JAVA
export JAVA_HOME=/usr/lib/jvm/java-8-openjdk-amd64
export PATH=$PATH:$JAVA_HOME/bin

# HADOOP
export HADOOP_HOME=/home/ubuntu/hadoop
export PATH=$PATH:$HADOOP_HOME/bin:$HADOOP_HOME/sbin


export HADOOP_CLASSPATH=$HADOOP_CLASSPATH:/home/ubuntu/hadoop/lib:/home/ubuntu/hadoop-lzo/target/hadoop-lzo-0.4.21-SNAPSHOT.jar
export JAVA_LIBRARY_PATH=/home/ubuntu/hadoop-lzo/target/native/Linux-amd64-64:/home/ubuntu/hadoop/lib/native

# GO
export GOROOT=/usr/local/go
export PATH=$GOPATH/bin:$GOROOT/bin:$PATH
```

### hadoop-env.sh (igual para master y slaves)
```bash
export JAVA_HOME=/usr/lib/jvm/java-8-openjdk-amd64/
export HADOOP_CLASSPATH=$HADOOP_CLASSPATH:/home/ubuntu/hadoop/lib:/home/ubuntu/hadoop-lzo/target/hadoop-lzo-0.4.21-SNAPSHOT.jar
export JAVA_LIBRARY_PATH=/home/ubuntu/hadoop-lzo/target/native/Linux-amd64-64:/home/ubuntu/hadoop/lib/native
export HADOOP_HOME=/home/ubuntu/hadoop
export HADOOP_HEAPSIZE_MAX=3000
export HADOOP_HEAPSIZE_MIN=2500
```
## 4 - Formatear master y slaves

Master:
```bash
rm -rf /tmp/hadoop*
sudo rm -rf /home/ubuntu/hadoop/hdfs/namenode
sudo mkdir -p /home/ubuntu/hadoop/hdfs/namenode
sudo chown -R ubuntu:ubuntu /home/ubuntu/hadoop/hdfs/namenode
~/hadoop/bin/hdfs namenode -format
```

Slaves
```bash
rm -rf /tmp/hadoop*
sudo rm -rf /home/ubuntu/hadoop/hdfs/datanode
sudo mkdir -p /home/ubuntu/hadoop/hdfs/datanode
sudo chown -R ubuntu:ubuntu /home/ubuntu/hadoop/hdfs/datanode
~/hadoop/bin/hdfs namenode -format
```

## 5 - Lanzar mapred y comprobar resultados
Master:
```bash
start-dfs.sh
start-yarn.sh

hdfs dfs -rm -r output.task4

/home/ubuntu/hadoop/bin/mapred streaming -input s3a://datasets.elasticmapreduce/ngrams/books/20090715/spa-all/1gram/data -output output.task4 -mapper "/home/ubuntu/task1 -task 0 -phase map" -reducer "/home/ubuntu/task1 -task 0 -phase reduce" -io typedbytes -inputformat SequenceFileInputFormat

hdfs dfs -cat output.task4/part-00000
```

NOTA: Hay veces que a hadoop le da por dejar procesos residuales del datanode en los slaves y cuando haces start-dfs.sh no te deja crearlos debido a eso (aunque no dice nada...). Para comprobar si esto está pasando, hacer un ps -axf | grep java y matar todos los procesos con kill -9 PID.

## Monitorización del job
```
http://52.22.18.120:8088/cluster
```
