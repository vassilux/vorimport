vorimport
=========

Import cdr asterisk datas from mysql backend to mongo


***** Installation
    Install mongodb and redis packages in the target system.
	
	Note : mongodb v 2.6.1 is used. Installation of packages is processing by avor_installation script.
	You can check this script for more information
    
    apt-get mongodb redis-server
    
    apt-get git mercurial bzr
    
    Execute make depends to install dependances of the project.
    
***** Configuration proxy
    Please pay a little attention for git proxy configuration.
    This is an example of the 
    
    export http_proxy=http://user:pass@proxyhost:proxyport
    git config --global http.proxy http://user:pass@proxyhost:proxyport

***** Requered packages
    Install debian packages
    
    apt-get install bzr
    
    apt-get install git mercurial

    Install go packages
    
    Mongodb driver :
    go get labix.org/v2/mgo
    
    MySql driver :
    go get github.com/ziutek/mymysql/thrsafe
    go get github.com/ziutek/mymysql/autorc
    go get github.com/ziutek/mymysql/godrv
    
    Seelog : 
    go get https://github.com/cihub/seelog
    
    Redis :
    cd $GOPATH/src
    git clone git://github.com/alphazero/Go-Redis.git redis
    cd redis
    go install
	
****** Application startup

	Application can be started with command line options
	
	--config , the json configuration file 
	
	--tick , the value on seconds for schedule the import task
	
****** Application tips

	Application can be stopped by a proper way , send kill [pid]. You can find the pid by ps aux|grep vorimport

  

    

