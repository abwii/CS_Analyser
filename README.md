# CS_Analyser

Plateforme distribuée d'analyse de démos Counter-Strike 2. Ce projet exploite une architecture microservices pour ingérer, traiter et restituer des données tactiques complexes issues de fichiers .dem.

## Architecture technique

* **Frontend :** React, Tailwind CSS
* **Backend API :** Go, GraphQL, WebSockets
* **Data Processing :** Rust, demoparser2
* **Infrastructure :** Kubernetes, Docker, AWS (S3, SQS)
* **Bases de données :** PostgreSQL (métadonnées), ClickHouse (séries temporelles et événements en jeu)

## Fonctionnalités principales

* **Ingestion haute performance :** Traitement asynchrone des fichiers de démo via une file de messages SQS et des workers Rust scalables.
* **Moteur statistique :** Calcul automatisé du rating HLTV 2.0 (KPR, DPR, KAST, Impact, ADR).
* **Profilage spatial :** Analyse de la télémétrie 3D (X, Y, Z) pour déterminer le positionnement dynamique et le rôle tactique des joueurs en phase défensive (CT).
* **Veto en direct :** Système de sélection et de bannissement de cartes en temps réel via WebSockets pour les formats BO1 et BO3.

## Structure du monorepo

* `/apps/web` : Application client (React)
* `/apps/api` : Serveur d'interface et de gestion de sessions (Go)
* `/apps/parser` : Service de déconstruction des binaires de jeu (Rust)
* `/infra` : Manifestes Kubernetes et scripts de déploiement

## Prérequis

* Docker et Docker Compose
* Go 1.21+
* Rust (cargo)
* Node.js 20+

## Installation locale

1. Cloner le dépôt.
2. Copier le fichier d'environnement : `cp .env.example .env`
3. Démarrer les services d'infrastructure (PostgreSQL, SQS local, Redis) via Docker : `docker-compose up -d`
4. Lancer l'API Go : `cd apps/api && go run main.go`
5. Lancer le worker Rust : `cd apps/parser && cargo run`
6. Lancer le frontend : `cd apps/web && npm install && npm run dev`

## Auteurs

Wassim Bacha
