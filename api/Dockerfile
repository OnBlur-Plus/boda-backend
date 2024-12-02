# Development Stage
FROM node:18-alpine as development

WORKDIR /app

COPY package*.json ./

RUN npm install

COPY . .

RUN npm run build

# Production Stage
FROM node:18-alpine as production

ARG NODE_ENV=production
ENV NODE_ENV=${NODE_ENV}

WORKDIR /app

COPY --from=development /app/dist ./dist
COPY --from=development /app/package.json .
COPY --from=development /app/package-lock.json .
COPY --from=development /app/.env .

RUN npm ci --only=production

EXPOSE 5000
CMD [ "npm", "run", "start:prod" ]