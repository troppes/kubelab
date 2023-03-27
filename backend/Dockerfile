FROM node:lts-alpine

WORKDIR /app
COPY . .
COPY .env_dist .env

RUN npm ci --only=production

EXPOSE 3000

CMD [ "sh", "entrypoint.sh" ]