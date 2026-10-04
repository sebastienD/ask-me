# ask-me

Petit jeu de questions/réponses en ligne de commande, écrit en Go, pour réviser du vocabulaire. Il est livré avec une liste de verbes irréguliers anglais ([anglais.csv](anglais.csv)).

## Principe

Le jeu lit un fichier CSV où chaque colonne est un « thème » (par exemple : base verbale, prétérit, participe passé, traduction) et chaque ligne un « sujet » (un verbe).

À chaque tour, pour chaque joueur :

1. une ligne est tirée au hasard ;
2. une colonne est donnée, une autre (différente) est demandée ;
3. le joueur tape sa réponse.

```
antoine, si la base verbale vaut arise, alors que vaut le prétérit ?
 👉 arose
👍
```

En cas d'erreur, la bonne réponse est affichée. À la fin de la partie, le joueur ayant le plus de bonnes réponses est désigné gagnant (en cas d'égalité, le premier joueur l'emporte), puis les résultats de chaque joueur sont affichés :

```
Le gagnant est...   alice

Résultats :
  alice  score  67 %  ✅ 2 réussites  ❌ 1 échec
  bob    score  33 %  ✅ 1 réussite  ❌ 2 échecs
```

Le score est le pourcentage de bonnes réponses, arrondi à l'entier le plus proche.

## Installation

Prérequis : Go 1.21 ou plus récent.

```bash
git clone https://github.com/sebastienD/ask-me.git
cd ask-me
go build
```

Des binaires sont aussi publiés dans les releases GitHub (avec le fichier `anglais.csv`).

## Utilisation

Le fichier `anglais.csv` doit se trouver dans le répertoire courant.

Partie solo (joueur par défaut : `antoine`) :

```bash
./ask-me
```

Partie à plusieurs, en local et au tour par tour — passer les prénoms en arguments :

```bash
./ask-me alice bob
```

Par défaut, chaque joueur répond à 3 questions. L'option `-n` change ce nombre ; elle doit être placée avant les noms des joueurs :

```bash
./ask-me -n 10 alice bob
```

## Format du fichier CSV

- séparateur : `;`
- la première ligne non vide est l'en-tête (les noms des thèmes) ;
- les lignes vides et celles commençant par `#` sont ignorées ;
- une ligne ayant moins de colonnes que l'en-tête est ignorée (avec un message dans les logs) ;
- il faut au moins 2 thèmes et 1 ligne valide ;
- les espaces autour des valeurs sont supprimés.

```csv
la base verbale;le prétérit;le participe passé;la traduction

# verbes en A
abide;abode;abode;respecter/se conformer à
arise;arose;arisen;survenir
```

La comparaison des réponses est exacte : casse, accents et variantes (`respecter/se conformer à`) doivent correspondre au caractère près. Seuls les espaces en début et fin de réponse sont ignorés.

## Développement

### Structure du code

| Fichier | Rôle |
|---|---|
| [main.go](main.go) | lecture des options et des joueurs, lancement de la partie |
| [deck.go](deck.go) | chargement et validation du fichier CSV (`LoadDeck`, `ParseDeck`), tirage des questions |
| [results.go](results.go) | joueurs, score, gagnant et affichage des résultats (communs à tous les modes) |
| [game.go](game.go) | partie dans le terminal : questions, réponses, enchaînement des tours |
| [room.go](room.go) | moteur de la partie en réseau : salon, modes de jeu, délais, diffusion de l'état aux joueurs |
| [snapshot.go](snapshot.go) | état de la partie envoyé aux joueurs (sans jamais révéler la réponse en cours) |

### Compiler

```bash
go build -v ./...
```

### Tests

```bash
go test -v ./...
```

Les tests unitaires couvrent :

- [deck_test.go](deck_test.go) : lecture du CSV (en-tête, commentaires, lignes vides, espaces, lignes trop courtes), cas d'erreur, et cohérence du fichier `anglais.csv` livré (aucune case vide), tirage des questions (deux thèmes toujours différents) ;
- [game_test.go](game_test.go) : vérification des réponses et enchaînement des tours entre joueurs ;
- [results_test.go](results_test.go) : désignation du gagnant (égalité : le premier joueur l'emporte), calcul du score et affichage des résultats ;
- [room_test.go](room_test.go) : partie en réseau — inscription (prénom vide, trop long ou déjà pris), déroulement des modes « chacun son tour » et « le plus rapide », temps écoulé, état envoyé aux joueurs.

Pour les tests, `Game` lit les réponses et écrit les questions via des champs remplaçables (`in`, `out`, `rng`) ; le générateur aléatoire est initialisé avec une graine fixe pour des tirages reproductibles. De la même façon, `Room` reçoit une fonction `schedule` : les tests la remplacent par une fausse horloge pour faire avancer le temps sans attendre.

La CI GitHub Actions ([.github/workflows/go.yml](.github/workflows/go.yml)) compile et lance les tests à chaque push et pull request sur `main`.

## Idées / TODO

- mode chrono (un maximum de bonnes réponses en un temps donné) ;
- mode « première lettre » ;
- multijoueur à distance (création de partie, statistiques, rejouer les erreurs…) ;
- accepter plusieurs traductions possibles ;
- tolérance sur les accents (paramétrable) ;
- prononciation des mots.

## Licence

Voir [LICENSE](LICENSE).
