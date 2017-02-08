FROM node
ADD /main.js main.js
ADD /package.json package.json
RUN npm install
ENTRYPOINT ["node", "main.js"]
