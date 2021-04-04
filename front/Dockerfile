FROM node:10.23.1

RUN mkdir -p /app
WORKDIR /app
COPY package.json /app
COPY package-lock.json /app
RUN npm install
COPY /src /app/src
COPY rollup.config.js /app/rollup.config.js
COPY /public /app/public
RUN rm -r /app/public/build

EXPOSE 5000
CMD ["npm", "run", "build-and-serve"]
