#  client-ag3 AG-3 (API GRILLE 3)

## Règles de base :

AG-3 est un jeu en temps réel où deux équipes, A et B, s'affrontent pour capturer le drapeau de l'équipe adverse. 

Les joueurs pourront choisir de rejoindre l'une des deux équipes, A ou B.

Chaque équipe dispose d'une grille, Grille A et Grille B, avec leur propre drapeau respectif. 

Les joueurs peuvent demander à l'API AG-3 une instance (appelée programme) qui agit comme un pointeur sur la grille.

Les programmes des joueurs disposent de 4 cellules, chacune ayant un niveau d'énergie qui diminue à chaque action (attaque, déplacement, réparation des cellules endommagées). 

Les joueurs doivent se déplacer sur une zone pour explorer les cellules correspondantes. Lorsqu'un programme est sur une zone, il peut capturer de l'énergie ou le drapeau si présent. 

Cependant, la capture laisse des traces (logs) qui peuvent être détectées par l'équipe adverse. 

Si l'ID du joueur est découvert, il peut être attaqué lors d'une exploration, et inversement.

Chaque cellule du programme peut explorer et capturer les cellules correspondantes de la zone (par exemple, la cellule 1 du programme peut explorer et capturer la cellule 1 de la zone). 

Les programmes peuvent également attaquer d'autres programmes. Pour lancer une attaque, les deux programmes doivent être en mode exploration chacun sur la cellule correspondante. 

Cependant, une attaque est possible que sur des cellules avec 100% de leur valeur.

Les joueurs peuvent nettoyer les logs pour éviter d'être détectés par l'équipe adverse. 

Ils peuvent également capturer des compétences pour améliorer leur programme. 

Les compétences peuvent être utilisées pour générer un shellcode qui sera envoyé à l'API AG-3. 

L'API retourne alors une liste des programmes avec leur ID et le shellcode généré. 

Si un joueur envoie une requête sur la zone avec le shellcode et l'ID correspondant, et que le programme qui a cet ID est sur la zone, il perd toutes ses cellules en exploration et leur valeur est réinitialisée à 0. 

Pour sortir du mode exploration, les cellules doivent être à 100% de leur valeur. Pour reconstruire les cellules et sortir de l'exploration, les joueurs ont besoin d'énergie.

Capture de Drapeau est un jeu de stratégie et d'action passionnant qui mettra à l'épreuve les compétences des joueurs en matière de défense et d'attaque. 

Rejoignez une équipe et participez à la bataille pour capturer le drapeau de l'équipe adverse !


### Equipe A

- un groupe ${\color{blue}BLUE TEAM}$ equipe A defend la grille A contre equipe B pour proteger flag A
- un groupe ${\color{red}RED TEAM}$ equipe A attaque la grille B pour capture flag B

### Equipe B

- un groupe ${\color{blue}BLUE TEAM}$ equipe B defend la grille B contre equipe A pour proteger flag B
- un groupe ${\color{red}RED TEAM}$ equipe B attaque la grille A pour capture flag A

### Usage 

```
> $ go build -o client-ag3 main.go      
```

```bash 
> $ ./client-ag3                                                                                                                                                                                                                                [±master ●●]

Usage: main [OPTIONS] COMMAND [arg...]

Client AG-3
                     
Options:             
  -v, --version      Show the version and exit
                     
Commands:            
  create             creation programme et chargement sur la grille
  load               charger programme existant sur la grille
  move               deplacer un programme sur la grille
  quick_move         [blue team] deplacer un programme sur la grille
  estimate_move      estimation temp deplacement sur zone
  stop_move          stopper navigation en cours ( retour zone de depart )
  scan               scan infos de la zone pour
  explore            exploration de cellule de zone
  destroy            destroy cellule programme
  rebuild            reconstruire cellule programme
  capture            capture data-energy cellule programme et zone
  equilibrium        repartir energie du programme uniformement
  pushflag           push drapeau dans zone de transfert
  status             status grille
  infos              infos programme
  navigation_stop    stop mode navigation
  exploration_stop   stop exploration
  monitoring         position + status programme monitoring
  log                info log cellule
  clean_log          clean log cellule
  destroy_zone       destroy cellule zone current
  shellcode          generate shellcode sur programmme present sur zone
  active_shellcode   activer un shellcode sur un programme
                     
Run 'main COMMAND --help' for more information on a command.
```
### Routes constanstes CLIENT

```
	API_URL_A                        = "http://195.154.84.18:20180"
	API_URL_B                        = "http://195.154.84.18:20190"
	ROUTE_NEW_PROGRAMME              = "v1/programme/new"
	ROUTE_LOAD_PROGRAMME             = "v1/programme/load"
	ROUTE_LOAD_PROGRAMME_BLUE_TEAM   = "v1/programme/load/blue/team"
	ROUTE_MOVE_PROGRAMME             = "v1/programme/move"
	ROUTE_QUICK_MOVE_PROGRAMME       = "v1/programme/quick/move"
	ROUTE_SCAN_PROGRAMME             = "v1/programme/scan"
	ROUTE_GET_CELLULE_LOG            = "v1/programme/cellule/log"
	ROUTE_CLEAN_CELLULE_LOG          = "v1/programme/clean/cellule/log"
	ROUTE_EXPLORE_PROGRAMME          = "v1/programme/explore"
	ROUTE_DESTROY_PROGRAMME          = "v1/programme/destroy"
	ROUTE_REBUILD_PROGRAMME          = "v1/programme/rebuild"
	ROUTE_STATUS_GRID                = "v1/grid"
	ROUTE_STATUS_PROGRAMME           = "v1/programme/infos"
	ROUTE_CAPTURE_CELL_DATA          = "v1/programme/capture/cellule/data"
	ROUTE_CAPTURE_CELL_ENERGY        = "v1/programme/capture/cellule/energy"
	ROUTE_CAPTURE_TARGET_DATA        = "v1/programme/capture/target/data"
	ROUTE_CAPTURE_TARGET_ENERGY      = "v1/programme/capture/target/energy"
	ROUTE_EQUILIBRiUM                = "v1/programme/equilibrium"
	ROUTE_PUSH_FLAG                  = "v1/programme/push/flag"
	ROUTE_DESTROY_ZONE               = "v1/programme/destroy/zone"
	ROUTE_NAVIGATION_PROGRAMME_STOP  = "v1/programme/navigation/stop"
	ROUTE_EXPLORATION_PROGRAMME_STOP = "v1/programme/exploration/stop"
	ROUTE_STOP_MOVE_PROGRAMME        = "v1/programme/stop/move"
	ROUTE_ESTIMATE_MOVE_PROGRAMME    = "v1/programme/estimate/move"
	ROUTE_ACTIVE_SHELLCODE           = "v1/active/shellcode"
	ROUTE_ZONE_SHELLCODE             = "v1/zone/shellcode"
```

### Params routes

```json
[
  {
    "path": "GET /v1/grid"
  },
  {
    "path": "GET /v1/active/capture/flag/:id/:secretid/:flag"
  },
  {
    "path": "GET /v1/programme/new/:name"
  },
  {
    "path": "GET /v1/programme/infos/:id/:secretid"
  },
  {
    "path": "POST /v1/programme/load"
  },
  {
    "path": "POST /v1/programme/load/blue/team"
  },
  {
    "path": "GET /v1/programme/move/:id/:secretid/:secteur_id/:zone_id"
  },
  {
    "path": "GET /v1/programme/quick/move/:id/:secretid/:secteur_id/:zone_id"
  },
  {
    "path": "GET /v1/programme/scan/:id/:secretid"
  },
  {
    "path": "GET /v1/programme/explore/:id/:secretid/:celluleid"
  },
  {
    "path": "GET /v1/programme/destroy/:id/:secretid/:celluleid/:targetid/:energy"
  },
  {
    "path": "GET /v1/programme/rebuild/:id/:secretid/:celluleid/:targetid/:energy"
  },
  {
    "path": "GET /v1/programme/capture/cellule/data/:id/:secretid/:celluleid/:index"
  },
  {
    "path": "GET /v1/programme/capture/cellule/energy/:id/:secretid/:celluleid/:index"
  },
  {
    "path": "GET /v1/programme/capture/target/data/:id/:secretid/:celluleid/:targetid"
  },
  {
    "path": "GET /v1/programme/capture/target/energy/:id/:secretid/:celluleid/:targetid"
  },
  {
    "path": "GET /v1/programme/equilibrium/:id/:secretid"
  },
  {
    "path": "GET /v1/programme/push/flag/:id/:secretid"
  },
  {
    "path": "GET /v1/programme/cellule/log/:id/:secretid/:celluleid"
  },
  {
    "path": "GET /v1/programme/clean/cellule/log/:id/:secretid/:celluleid"
  },
  {
    "path": "GET /v1/programme/destroy/zone/:id/:secretid/:celluleid/:energy"
  },
  {
    "path": "GET /v1/programme/navigation/stop/:id/:secretid"
  },
  {
    "path": "GET /v1/programme/exploration/stop/:id/:secretid"
  },
  {
    "path": "GET /v1/programme/stop/move/:id/:secretid"
  },
  {
    "path": "GET /v1/programme/estimate/move/:id/:secretid/:secteur_id/:zone_id"
  },
  {
    "path": "GET /v1/zone/shellcode/:id/:secretid"
  },
  {
    "path": "GET /v1/active/shellcode/:id/:secretid/:target_id/:shellcode",
  }
]
```