FROM node:20-bullseye-slim AS base

WORKDIR /app

COPY . .

CMD [ "npm", "rum", "serve" ]