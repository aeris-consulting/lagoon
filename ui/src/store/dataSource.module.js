import {
    SET_DATASOURCE,
    FETCH_DATASOURCE
} from './actions.type';

import { DatasourcesService } from '../services/api.service'

const initialState = {
    datasources: [],
}

const state = { ...initialState }

const getters = {
    getDataSourceById: (state) => (id) => {
        return state.dataSources.find(dataSource => dataSource.id === id)
    }
}

export const actions = {
    async [FETCH_DATASOURCE](context) {
        const { data } = await DatasourcesService.getDataSources();
        context.commit(SET_DATASOURCE, data.datasources);
        return data.datasources;
    }
}

export const mutations = {
    [SET_DATASOURCE](state, datasources) {
        state.datasources = datasources;
    },
}

export default {
    state,
    actions,
    mutations,
    getters
}