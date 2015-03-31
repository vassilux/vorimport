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
Addapter ce fichier à la configuraiton de système : le dialplan de iPBX et les besoins du client.

Paramètre cleanupRequests permet d'exécuter les requetes personnalisées avant lencement d'importaiton de données de mysql vers mongo.
Mot clé delete est interdit à utiliser.

Paramètre excludeFromAnalytics permet exclure des SDA de processus d'importation. Ce cas peut être intéressant car les numéro des SDA sont équivalant au numéro des postes.

Importat : Importation de SDA est basée sur les SDA saisie dans la base de données. Ces SDA peut être saisie via l'interface graphique

Context app-alive-test de asterisk dialplan est utilisé pour générer un test cyclique de toute la chaine
Ce context peut être ajouter dans /etc/asterisk/extensions_custom.conf.
Voici un example
	
		[app-alive-test]
		exten => testcall,1,NoOp(Process test call)
		same => n,Answer()
		exten => h,1,ResetCDR()
		same => n,NoCDR()
		
Merci de créer un utilisateur vorimport(par defaut) dans /etc/asterisk/manager.conf et adapter les droits au système cible si
la fonction de test d'appel est activée.

		[vorimport]
		secret = crackme
		deny=0.0.0.0/0.0.0.0
		permit=192.168.3.0/255.255.255.0
		permit=127.0.0.1/255.255.255.0
		read = system,call,log,verbose,agent,user,config,dtmf,reporting,cdr,dialplan
		write = system,call,agent,user,command,config,reporting,originate,message

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





