const DEFAULT_HOST = "http://localhost";
const DEFAULT_PORT = "8080";

const host = __process.env["API_HOST"] || DEFAULT_HOST;
const port = __process.env["API_PORT"] || DEFAULT_PORT;

export const targetHost = `${host}:${port}`;
console.log(`Targetting API at: "${targetHost}"`);