import Vue from 'vue'
import Vuex from 'vuex'

import datasource from './datasource.module'
import error from './error.module'

Vue.use(Vuex)

export default new Vuex.Store({
  modules: {
    datasource,
    error
  }
})