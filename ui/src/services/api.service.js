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
  },

  get(url, params) {
    return Vue.axios.get(url, params).catch(error => {
      throw new Error(`Query Error ApiService ${error}`);
    });
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
    return Vue.axios.delete(resource).catch(error => {
      throw new Error(`Delete Error ApiService ${error}`);
    });
  }
};

export const DatasourcesService = {
  getDatasources() {
    return ApiService.get('datasource')
  },

  listEntryPoints(requestObj) {
    const { id, filter, minLevel, maxLevel } = requestObj;
    return ApiService.get(`data/${id}/entrypoint`, {
      params: {
        filter,
        min: minLevel,
        max: maxLevel,
      }
    })
  },

  getEntryPointsFromWebsocket(link) {
    let receivedValues = [];
    return new Promise((resolve, reject) => {
      let socket = new WebSocket(wsRoot + link);
      socket.onopen = () => {
        socket.onmessage = ({ data }) => {
          let jsonData = JSON.parse(data);
          if (jsonData.size > 0) {
            receivedValues = receivedValues.concat(jsonData.data);
          } else {
            // eslint-disable-next-line
            console.log("Closing the websocket");
            socket.close(1000, "End of data");
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