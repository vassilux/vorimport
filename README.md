vorimport
=========

Import cdr asterisk datas from mysql backend to mongo


***** Installation
    Install mongodb and redis packages in the target system
    
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

  

    

