FROM node:10.23.1

# build static content
RUN mkdir -p /usr/src/app
WORKDIR /usr/src/app
COPY package.json /usr/src/app
COPY package-lock.json /usr/src/app
RUN npm install
COPY . /usr/src/app
RUN npm run build

# serve static content
EXPOSE 5000
CMD ["npm", "run", "start"]
