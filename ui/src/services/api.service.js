import axios from 'axios'
import Vue from "vue";
import VueAxios from "vue-axios";
import _ from 'lodash';

let apiRoot;
let wsRoot;

if (!_.isNil(process) && !_.isNil(process.env) && !_.isNil(process.env.VUE_APP_API_BASE_URL) && !_.isNil(process.env.VUE_APP_WS_BASE_URL)) {
  apiRoot = process.env.VUE_APP_API_BASE_URL;
  wsRoot = process.env.VUE_APP_WS_BASE_URL;
} else {
  apiRoot = location.pathname + '..';
  if (location.protocol == 'https:') {
    wsRoot = 'wss://';
  } else {
    wsRoot = 'ws://';
  }
  wsRoot += location.hostname + ':' + location.port + apiRoot
}

export const ApiService = {
  init() {
    Vue.use(VueAxios, axios);
    Vue.axios.defaults.baseURL = apiRoot;
    Vue.axios.interceptors.response.use((response) => {
      return response;
    }, (error) => {
      if (error && error.response && error.response.status) {
        if (error.response.status === 401 || error.response.status === 404) {
          location.reload();
        }
      }
      return Promise.reject(error);
    })
  },

  get(url, params) {
    return Vue.axios.get(url, params);
  },

  post(url, params) {
    return Vue.axios.post(`${url}`, params);
  },

  update(url, params) {
    return Vue.axios.put(`${url}`, params);
  },

  put(url, params) {
    return Vue.axios.put(`${url}`, params);
  },

  delete(resource) {
    return Vue.axios.delete(resource);
  },

};

export const DatasourcesService = {
  async executeCommand(commands, nodeId, datasourceId) {
    return ApiService.post(`data/${encodeURIComponent(datasourceId)}/command`, {args: commands, nodeId: nodeId})
        .then(response => {
          return response.data;
        }).catch(e => {
          return Promise.reject(e.response.data.error)
        });
  },

  getClusterNodes(datasourceId) {
    return ApiService.get(`data/${encodeURIComponent(datasourceId)}/infos`)
        .then(response => {
          const infos = response.data.infos;
          if (infos != null && infos.nodes != null) {
            return infos.nodes;
          }
          return [];
        }).catch(e => {
          return Promise.reject(e.response.data.error)
        })
  },

  getNodeDetails(datasource, node) {
    const fullPath = node.fullPath;
    const nodeResourcePath = `data/${encodeURIComponent(datasource.id)}/entrypoint/${encodeURIComponent(fullPath)}`
    const details = {};
    return new Promise((resolve, reject) => {
      ApiService.get(`${nodeResourcePath}/info`, {format: 'json'})
        .then(response => {
          details.info = response.data;
          ApiService.get(`${nodeResourcePath}/content`, {format: 'json'})
              .then(response => {
                if (response.status === 200) {
                  details.content = response.data
                  resolve(details)
                } else if (response.status === 202) {
                  let receivedValues = [];
                  let socket = new WebSocket(this.wsRoot + response.data.link);
                  socket.onopen = () => {
                    socket.onmessage = ({data}) => {
                      let jsonData = JSON.parse(data);
                      if (jsonData.size) {
                        receivedValues = receivedValues.concat(jsonData.data);
                      } else {
                        details.content = {
                          length: receivedValues.length,
                          data: receivedValues
                        };
                        resolve(details);
                      }
                  };
                };
              }
            })
            .catch(e => {
              reject(e.response.data.error)
            })
        }).catch(e => {
          reject(e)
        })
    });
  },

  getDatasources() {
    return ApiService.get('datasource').then(response => response.data)
  },

  listEntryPoints(requestObj) {
    const { id, filter, minLevel, maxLevel } = requestObj;
    return ApiService.get(`data/${encodeURIComponent(id)}/entrypoint`, {
      params: {
        filter,
        min: minLevel,
        max: maxLevel,
      }
    }).catch(e => {
      return Promise.reject(e.response.data.error)
    })
  },

  deleteEntrypoint(datasourceId, fullPath) {
    return ApiService.delete(`data/${encodeURIComponent(datasourceId)}/entrypoint/${encodeURIComponent(fullPath)}`, {format: 'json'})
        .catch(e => {
          return Promise.reject(e.response.data.error)
        })
  },

  deleteEntrypointChildren(datasourceId, fullPath) {
    return ApiService.delete(`data/${encodeURIComponent(datasourceId)}/entrypoint/${encodeURIComponent(fullPath)}/children`, {format: 'json'})
        .catch(e => {
          return Promise.reject(e.response.data.error)
        })
  },

  getEntryPointsFromWebsocket(link) {
    let receivedValues = [];
    return new Promise((resolve, reject) => {
      let socket = new WebSocket(wsRoot + link);
      socket.onopen = () => {
        socket.onmessage = ({data}) => {
          let jsonData = JSON.parse(data);
          if (jsonData.size > 0) {
            receivedValues = receivedValues.concat(jsonData.data);
          } else {
            // eslint-disable-next-line
            console.log("Closing the websocket");
            socket.close(1000, "End of data");
            receivedValues = receivedValues
              .sort((a, b) => {
                return a.path < b.path ? -1 : 1
            });
            resolve(receivedValues);
          }
        };
        socket.onerror = (e) => {
          // eslint-disable-next-line
          console.log("The websocket got an error: ", e);
          reject(e);
        };
        socket.onclose = () => {
          // eslint-disable-next-line
          console.log("The websocket is closed");
        };
      }
    })
  }
}