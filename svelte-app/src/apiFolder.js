import axios from 'axios';

const hostname = "http://localhost:8080";
const foldersUrl = `${hostname}/folders`;

export const apiGetFolderContent = (folderId = undefined) => {
  return new Promise(resolve => {
    const folderUrl = folderId ? `${foldersUrl}/${folderId}` : foldersUrl;
    axios({ method: "GET", url: folderUrl }).then(response => {
      resolve(response.data);
    });
  });
};

export const apiUpdateFolder = ({ id, name, parentId }) => {
  if (id === undefined) {
    console.error("need an id for the folder to update");
    return;
  }

  return new Promise(resolve => {
    axios({
      method: "PUT",
      url: `${foldersUrl}/${id}`,
      data: { id, name, parentId }
    }).then(() => resolve());
  });
};

export const apiDeleteFolder = id => {
  if (id === undefined) {
    console.error("need an id for the folder to edit");
    return;
  }

  return new Promise(resolve => {
    axios({ method: "DELETE", url: `${foldersUrl}/${id}` }).then(response => {
      resolve();
    });
  });
};

export const apiMoveFolder = (folderId, destFolderId) => {
  return new Promise(resolve => {
    axios({
      method: "PUT",
      url: `${hostname}/MoveFolder/${folderId}?dest=${destFolderId}`
    }).then(response => {
      resolve();
    });
  });
};

export const apiCreateFolder = ({ name, parentId }) => {
  return new Promise(resolve => {
    axios({
      method: "POST",
      url: `${foldersUrl}`,
      data: { name, parentId }
    }).then(response => {
      resolve(response.data);
    });
  });
};
