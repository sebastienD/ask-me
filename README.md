# ask-me

Petit jeu de questions/réponses pour réviser du vocabulaire, seul ou entre amis. Il est livré avec la liste des verbes irréguliers anglais ([anglais.csv](anglais.csv)), mais fonctionne avec n'importe quelle liste.

Trois façons de jouer :

| | Dans le terminal | Sur téléphone, même Wi-Fi | Sur téléphone, à distance |
|---|---|---|---|
| Pour qui | seul, ou à plusieurs sur le même ordinateur | à plusieurs, réunis au même endroit | à plusieurs, chacun chez soi |
| Comment | on se passe le clavier | une personne crée la partie, les autres rejoignent avec un lien | pareil, le lien marche aussi en 4G/5G |
| À installer | le jeu sur l'ordinateur | le jeu sur un seul ordinateur, **rien sur les téléphones** | le jeu et [cloudflared](#jouer-à-distance) sur un seul ordinateur |
| Commande | `./ask-me` | `./ask-me -local` | `./ask-me -remote` |

## Le principe

Chaque ligne de la liste est un verbe, chaque colonne une de ses formes (base verbale, prétérit, participe passé, traduction). Pour chaque question, le jeu tire un verbe au hasard, donne une de ses formes et en demande une autre :

```
alice, si la base verbale vaut arise, alors que vaut le prétérit ?
 👉 arose
👍
```

En cas d'erreur, la bonne réponse est affichée.

## Installation

### Télécharger le jeu (sans rien installer d'autre)

1. Ouvrir la [dernière release](https://github.com/sebastienD/ask-me/releases/latest) et télécharger l'archive de son système :

   | Système | Archive à prendre |
   |---|---|
   | Mac avec puce Apple (M1, M2, M3, M4…) | `ask-me-…-darwin-arm64.tar.gz` |
   | Mac avec processeur Intel | `ask-me-…-darwin-amd64.tar.gz` |
   | Windows | `ask-me-…-windows-amd64.zip` |
   | Linux | `ask-me-…-linux-amd64.tar.gz` (ou `linux-arm64`, `linux-386`) |

   **Sur Mac, « amd64 » veut dire Intel**, pas « version standard ». Pour savoir quelle puce a son Mac : menu  › À propos de ce Mac, ligne « Puce » (Apple M…) ou « Processeur » (Intel). Prendre la mauvaise archive donne l'erreur « bad CPU type in executable ».

2. Décompresser l'archive (double-clic). Le dossier obtenu contient le jeu, `anglais.csv`, ce README et la licence.

3. Ouvrir un terminal **dans ce dossier** et lancer le jeu :

   ```bash
   ./ask-me
   ```

   Sous Windows, dans PowerShell : `.\ask-me.exe`.

### Premier lancement : l'avertissement de sécurité

Le jeu n'est pas signé auprès d'Apple ni de Microsoft : au premier lancement, le système le bloque, comme pour tout programme téléchargé dont il ne connaît pas l'auteur. C'est normal, il suffit de l'autoriser **une seule fois**.

**Sur Mac**, le message est « Apple n'a pas pu confirmer que « ask-me » ne contenait pas de logiciel malveillant… ». Au choix :

- dans le terminal, depuis le dossier du jeu :

  ```bash
  xattr -d com.apple.quarantine ask-me
  ```

- ou sans terminal : fermer le message, puis  › Réglages Système › Confidentialité et sécurité ; en bas, à côté de « ask-me a été bloqué », cliquer sur **Ouvrir quand même** et confirmer.

**Sur Windows**, si le message « Windows a protégé votre ordinateur » s'affiche : cliquer sur **Informations complémentaires**, puis **Exécuter quand même**. Dans PowerShell, on peut aussi débloquer le fichier avant de le lancer :

```powershell
Unblock-File .\ask-me.exe
```

### Compiler le jeu soi-même

Il faut [Go](https://go.dev/dl/) 1.21 ou plus récent :

```bash
git clone https://github.com/sebastienD/ask-me.git
cd ask-me
go build
```

Un jeu compilé sur sa propre machine n'est pas concerné par l'avertissement de sécurité.

Dans tous les cas, le fichier `anglais.csv` doit se trouver dans le dossier depuis lequel on lance le jeu.

## Jouer dans le terminal

Seul (le joueur s'appelle alors `antoine`) :

```bash
./ask-me
```

À plusieurs : passer les prénoms. Les joueurs répondent chacun leur tour.

```bash
./ask-me alice bob
```

Chaque joueur répond à 3 questions par défaut. L'option `-n` change ce nombre ; elle doit être placée **avant** les prénoms :

```bash
./ask-me -n 10 alice bob
```

## Jouer sur téléphone, en réseau local

### 1. Créer la partie

Sur un ordinateur connecté au Wi-Fi :

```bash
./ask-me -local
```

Le jeu demande d'abord le mode de jeu (voir [Les modes de jeu](#les-modes-de-jeu)) :

```
Mode de jeu :
  1) Chacun son tour
  2) Le plus rapide
Ton choix [1] :
```

Il affiche ensuite le lien de la partie, un message prêt à copier dans WhatsApp et un QR code :

```
🎮 Partie créée (mode : Le plus rapide) !

📱 Pour jouer, ouvrez ce lien sur votre téléphone (toi aussi !) :

   http://192.168.1.23:4242

💬 Message à copier dans WhatsApp :

   Viens jouer à ask-me avec moi 🎮 👉 http://192.168.1.23:4242
```

### 2. Rejoindre

Chacun touche le lien reçu sur WhatsApp, ou scanne le QR code, puis tape son prénom. **Celui qui a créé la partie ouvre le même lien pour jouer aussi.** Le terminal affiche les arrivées :

```
✅ Léa a rejoint la partie (1 joueur)
✅ Tom a rejoint la partie (2 joueurs)
```

### 3. Jouer

Quand tout le monde est là, appuyer sur **Entrée** dans le terminal. Sur les téléphones :

- la question s'affiche avec un compte à rebours de **30 secondes** ;
- puis la bonne réponse et le nom de celui qui l'a trouvée, pendant 3 secondes ;
- le tableau des scores est mis à jour en direct ;
- à la fin, le gagnant et les résultats s'affichent (aussi dans le terminal).

Pour rejouer, appuyer sur Entrée dans le terminal pour quitter, puis relancer `./ask-me -local` et renvoyer le nouveau lien.

### Options

| Option | Rôle | Défaut |
|---|---|---|
| `-n` | nombre de questions par joueur (en mode « le plus rapide » : nombre total de questions) | 3 |
| `-port` | port utilisé par la partie (à changer si 4242 est déjà pris) | 4242 |
| `-remote` | jouer à distance, par Internet, au lieu du même Wi-Fi (voir [Jouer à distance](#jouer-à-distance)) | désactivé |

```bash
./ask-me -local -n 10
```

## Jouer à distance

Pour jouer avec des amis qui ne sont pas sur le même Wi-Fi (chacun chez soi, ou en 4G/5G), la partie passe par Internet grâce à un tunnel Cloudflare : gratuit, sans compte, et sans rien configurer sur la box.

### Installer cloudflared (une seule fois, sur l'ordinateur de l'hôte)

| Système | Commande |
|---|---|
| macOS | `brew install cloudflared` |
| Windows | `winget install --id Cloudflare.cloudflared` |
| Linux | voir [la page de téléchargement](https://developers.cloudflare.com/cloudflare-one/connections/connect-networks/downloads/) |

### Créer la partie

```bash
./ask-me -remote
```

Tout se passe ensuite comme [sur le même Wi-Fi](#jouer-sur-téléphone-en-réseau-local) : choix du mode, lien à envoyer sur WhatsApp, QR code, **Entrée** pour lancer. Seul le lien change, il ressemble à :

```
https://vert-lapin-42.trycloudflare.com
```

- Le jeu attend que le lien réponde avant de l'afficher (quelques secondes).
- Le lien change à chaque partie : il faut renvoyer le nouveau.
- **L'ordinateur de l'hôte doit rester allumé et connecté** pendant toute la partie.
- Les options `-n` et `-port` fonctionnent de la même façon.

En mode « le plus rapide », c'est la première réponse **reçue** qui gagne : à distance, un joueur avec une connexion lente peut perdre quelques dixièmes de seconde.

## Les modes de jeu

Ces modes concernent la partie sur téléphone. Dans le terminal, on joue toujours chacun son tour.

| | Chacun son tour | Le plus rapide |
|---|---|---|
| Questions | chaque joueur a sa propre question, à tour de rôle | tout le monde a la même question en même temps |
| Qui marque | le joueur s'il trouve | le **premier** qui trouve |
| Mauvaise réponse | compte comme un échec | compte comme un échec, et on ne peut plus retenter cette question |
| Temps écoulé (30 s) | compte comme un échec | personne n'est pénalisé |
| Nombre de questions | `-n` par joueur | `-n` au total |

## Score et gagnant

À la fin de la partie, chaque joueur voit son score, ses réussites et ses échecs :

```
Résultats :
  Léa  score  67 %  ✅ 2 réussites  ❌ 1 échec
  Tom  score  33 %  ✅ 1 réussite  ❌ 2 échecs
```

- le **score** est le pourcentage de réussites parmi les réussites et les échecs du joueur, arrondi à l'entier (en mode « le plus rapide », une question à laquelle on n'a pas répondu ne compte pas) ;
- le **gagnant** est celui qui a le plus de réussites ; en cas d'égalité, c'est le premier inscrit.

## Comment les réponses sont vérifiées

La réponse doit être **exactement** celle de la liste : majuscules, accents et barres obliques comptent. Seuls les espaces au début et à la fin sont ignorés.

- `held` est juste pour le prétérit de `hold`, `Held` est faux ;
- quand la liste propose plusieurs formes, comme `burnt/burned` ou `respecter/se conformer à`, il faut taper la ligne entière.

Sur téléphone, la majuscule automatique et le correcteur sont désactivés dans le champ de réponse.

## Utiliser sa propre liste

Pour réviser autre chose, remplacer le contenu de `anglais.csv` (le nom du fichier est fixe pour l'instant). Le format :

```csv
la base verbale;le prétérit;le participe passé;la traduction

# verbes en A
abide;abode;abode;respecter/se conformer à
arise;arose;arisen;survenir
```

- le séparateur est `;` ;
- la première ligne donne le nom des colonnes ; il en faut au moins 2 ;
- les lignes vides et celles qui commencent par `#` sont ignorées ;
- une ligne avec moins de colonnes que la première est ignorée (un message l'indique) ;
- les espaces autour des valeurs sont supprimés.

## Questions fréquentes

**« bad CPU type in executable » au lancement sur Mac.**
L'archive téléchargée ne correspond pas à la puce du Mac : prendre `darwin-arm64` pour un Mac avec puce Apple (M1, M2…), `darwin-amd64` pour un Mac Intel. Voir [Télécharger le jeu](#télécharger-le-jeu-sans-rien-installer-dautre).

**« Apple n'a pas pu confirmer que « ask-me » ne contenait pas de logiciel malveillant ».**
Le jeu n'est pas signé auprès d'Apple : il faut l'autoriser une fois. Voir [Premier lancement](#premier-lancement--lavertissement-de-sécurité).

**« Windows a protégé votre ordinateur ».**
Même raison : cliquer sur « Informations complémentaires » puis « Exécuter quand même ».

**Le lien ne s'ouvre pas sur le téléphone.**
Avec `-local`, le téléphone doit être sur **le même Wi-Fi** que l'ordinateur, pas en 4G/5G. Pour jouer sans être sur le même Wi-Fi, lancer la partie avec [`-remote`](#jouer-à-distance).

**Tout le monde est sur le même Wi-Fi, mais ça ne marche toujours pas.**
Certains Wi-Fi « invités » ou d'établissements scolaires empêchent les appareils de se parler entre eux. Utiliser le Wi-Fi de la maison, ou un partage de connexion depuis un téléphone (l'ordinateur et les autres téléphones s'y connectent).

**macOS demande s'il faut autoriser les connexions entrantes.**
Cliquer sur « Autoriser », sinon les téléphones ne peuvent pas rejoindre la partie. Avec `-remote`, la question ne se pose pas : seul cloudflared se connecte au jeu.

**Le terminal affiche « Impossible de trouver l'adresse de cet ordinateur sur le réseau local ».**
L'ordinateur n'est connecté à aucun réseau : le connecter au Wi-Fi puis relancer la partie.

**« cloudflared est introuvable ».**
`-remote` a besoin de cloudflared sur l'ordinateur de l'hôte : voir [Installer cloudflared](#installer-cloudflared-une-seule-fois-sur-lordinateur-de-lhôte).

**Avec `-remote`, le lien ne s'ouvre pas tout de suite.**
Un nouveau lien met parfois quelques secondes à fonctionner partout. Le jeu prévient s'il ne répond pas encore : attendre un peu et réessayer.

**Un joueur a fermé la page ou son téléphone s'est mis en veille.**
Il rouvre le lien sur le même téléphone et retrouve sa place. Pendant ce temps, en mode « chacun son tour », ses questions sans réponse comptent comme des échecs.

**Quelqu'un arrive après le lancement.**
Il ne peut plus s'inscrire, mais il voit la partie en spectateur. Il pourra jouer à la suivante.

**« Ce prénom est déjà pris ».**
Deux joueurs ne peuvent pas avoir le même prénom (les majuscules ne comptent pas) : ajouter une initiale, par exemple « Léa B. ».

**Ma réponse était bonne mais elle est comptée fausse.**
Voir [Comment les réponses sont vérifiées](#comment-les-réponses-sont-vérifiées) : il faut la forme exacte de la liste, avec toutes ses variantes.

## Développement

```bash
go build ./...
```

```bash
go test -race ./...
```

L'architecture, le moteur de la partie en réseau, l'API HTTP et la stratégie de tests sont décrits dans la [documentation technique](docs/ARCHITECTURE.md). La CI GitHub Actions compile et lance les tests à chaque push et pull request sur `main`.

## Idées

- mode chrono (un maximum de bonnes réponses en un temps donné) ;
- mode « première lettre » ;
- statistiques de fin de partie (le plus rapide, rejouer les erreurs…) ;
- accepter une seule des variantes (`burnt` ou `burned`) ;
- tolérance sur les accents (paramétrable) ;
- choisir le fichier de vocabulaire ;
- prononciation des mots.

## Licence

Voir [LICENSE](LICENSE).
