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

En cas d'erreur, la bonne réponse est affichée. À la fin de la partie, le joueur ayant le plus de bonnes réponses est désigné gagnant.

## Installation

Prérequis : Go 1.21 ou plus récent.

```bash
git clone <url-du-depot> ask-me
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
- les espaces autour des valeurs sont supprimés.

```csv
la base verbale;le prétérit;le participe passé;la traduction

# verbes en A
abide;abode;abode;respecter/se conformer à
arise;arose;arisen;survenir
```

La comparaison des réponses est exacte : casse, accents et variantes (`respecter/se conformer à`) doivent correspondre au caractère près.

## Développement

```bash
go build -v ./...
```

```bash
go test -v ./...
```

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
