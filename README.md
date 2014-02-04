vorimport
=========

Import cdr asterisk datas from mysql backend to mongo

***** Requered packages
    seelog   : https://github.com/cihub/seelog
    redis    : https://github.com/alphazero/Go-Redis
    mymysql  : https://github.com/ziutek/mymysql/thrsafe
             : https://github.com/ziutek/mymysql/autorc
             : https://github.com/ziutek/mymysql/godrv
    mgo      : labix.org/v2/mgo

***** Installation
    Please check if redis server and mongo server are installed and configured on the target system.
    Execute make depends to install dependances of the project.
    
***** Configuration proxy
    Please pay a little attention for git proxy configuration.
    This is an example of the  
    export http_proxy=http://user:pass@proxyhost:proxyport
    git config --global http.proxy http://user:pass@proxyhost:proxyport
    

