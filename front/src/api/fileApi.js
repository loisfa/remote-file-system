import axios from "axios";
import { targetHost } from "./constants.js";

const filesUrl = `${targetHost}/files`;

export const getDownloadUrl = (file) => `${targetHost}/DownloadFile/${file.id}`;

export const apiDeleteFile = id => {
  if (id === undefined) {
    console.error("need an id for the file to edit");
    return;
  }

  return new Promise(resolve => {
    axios({ method: "DELETE", url: `${filesUrl}/${id}` }).then(response => {
      resolve();
    });
  });
};

export const apiMoveFile = (fileId, destFileId) => {
  return new Promise(resolve => {
    axios({
      method: "PUT",
      url: `${targetHost}/MoveFile/${fileId}?dest=${destFileId}`
    }).then(response => {
      resolve();
    });
  });
};

export const apiUploadFile = (jsFile, destFolderId) => {
  return new Promise(resolve => {
    const dest = destFolderId || 0;
    const data = new FormData();
    data.append('file', jsFile, jsFile.fileName);
    axios({
      method: "POST",
      data,
      url: `${targetHost}/UploadFile?dest=${destFolderId}`
    }).then(response => {
      resolve(response.data);
    });
  });
};
