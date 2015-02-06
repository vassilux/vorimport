----------
vorimport 
----------

 
### Installation 
Copier le ficheir vorimport_[version].tar.gz sur la machine cible dans le repertoire /opt/vorimport
Exécuter tar xvzf vorimport_[version].tar.gz

Ce placer dans le repertoire : cd /opt/vorimport/vorimport_[version]

> **NOTE:**
>
> - En cas d'une nouvelle installation copier le ficheir vorimport.supervisor.conf dans le répertoire /etc/supervisor/conf.d
>
> - Copier le fichier /opt/vorimport/vorimport_[version]/config.sample.json dans /opt/vorimport/vorimport_[version]/config.json
> - Addapter ce fichier en cas de besoin à votre environement cible


Vérifier si l'application est en cours d'execution via console de supervisor : supervisorctl
Si l'application est en court d'exécution arrêter l'application : stop vorimport
Quitter le console : exit

Crée un lien symbolic ln -s /opt/revor/vorimport_[version] /opt/vorimport/current 
> **NOTE:**
>
> - En cas si le répertoire existe /opt/vorimport/current. Supprimer rm -rf /opt/vorimport/current

### Configuraiton
Copier le fichier config.sample.json en config.json : cp config.sample.json config.json
Addapter ce fichier à la configuraiton de système : le dialplan de iPBX et les besoin du client.

Context app-alive-test de asterisk dialplan est utilisé pour générer un test cyclique de toute la chaine
Ce context peut être ajouter dans /etc/asterisk/extensions_custom.conf.
Voici un example
	
		[app-alive-test]
		exten => testcall,1,NoOp(Process test call)
		same => n,Answer()
		exten => h,1,ResetCDR()
		same => n,NoCDR()
		
Seule chose d'imposée d'avoir une extention <b>testcall</b>.

### Mise à jour 
Mise à jour est identique à l'installation sans la partie de la configuration.

Il faut copier le ficheir de la configuration actuel (/opt/vorimport/current/config.json) dans le repertoire /opt/vorimport/vorimport_[version] 


## Notes
		Cluster
		code NODSTO : L'information complimentaire Cluster resource stopped on astnode1
		code NODSTA : L'information complimentaire resource started on astnode1

		Statistiques:
		APPSTA  : application vorimport démarrée
		APPSTO  : application vorimport arrêtée
		MYSQKO  : connexion à la base de données mysql ok
		MYSQOK  : connexion à la base de données mysql ko
		MONGKO  : connexion à la base de données mongo ok
		MONGOK  : connexion à la base de données mongo ko
		TCALOK  : connexion au serveur astersik(l'application) est ok
		TCALKO  : connexion au serveur astersik(l'application) est ko
		CCALOK  : generation d'un appel de test ok 
		CCALKO  : generation d'un appel de test ok 





