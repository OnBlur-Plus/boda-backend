FROM node:18-alpine AS builder
COPY package*.json ./
RUN npm install
COPY . ./
RUN npm run build

FROM node:18-alpine AS node_modules
COPY package*.json ./
RUN npm install --production

FROM node:18-alpine as production
COPY --from=node_modules /node_modules ./node_modules
COPY --from=builder /package*.json ./
COPY --from=builder /dist ./dist
COPY --from=builder /.env ./.env

EXPOSE 5000
CMD [ "npm", "run", "start:prod" ]