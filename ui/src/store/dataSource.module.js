import {
    SET_DATASOURCE,
    FETCH_DATASOURCE,
    FETCH_ENTRY_POINTS
} from './actions.type';

import { DatasourcesService } from '../services/api.service'

const initialState = {
    datasources: [],
}

const state = { ...initialState }

const getters = {
    getDataSourceById: (state) => (id) => {
        return state.datasources.find(datasource => datasource.id === id)
    }
}

export const actions = {
    async [FETCH_DATASOURCE](context) {
        const { data } = await DatasourcesService.getDatasources();
        context.commit(SET_DATASOURCE, data.datasources);
        return data.datasources;
    },
    async [FETCH_ENTRY_POINTS](context, request) {
        const {id, filter, entrypointPrefix, minLevel, maxLevel} = request;
        let actualFilter;
        let overallFilter = ('*' + filter + '*').replace(/[*]+/g, '*');

        if (!entrypointPrefix) {
            actualFilter = ('*' + filter + '*').replace(/[*]+/g, '*');
        } else {
            let entrypointRegex = new RegExp(entrypointPrefix);
            entrypointPrefix = entrypointPrefix + ':*';
            let overallRegex = new RegExp(overallFilter
                .replace(/^[*]+/g, '')
                .replace(/[*]+$/g, '')
                .replace(/[*]+/g, '.*')
            );

            if (overallRegex.test(entrypointPrefix)) {
                actualFilter = entrypointPrefix;
            } else if (entrypointRegex.test(overallFilter
                .replace(/^[*]+/g, '')
                .replace(/[*]+$/g, '')
            )) {
                actualFilter = overallFilter;
            } else {
                actualFilter = overallFilter + ','
                    + entrypointPrefix
                        .replace(/[*]+/g, '.*')
                        .replace(/[*]+/g, '*');
            }
        }

        const response = await DatasourcesService.listEntryPoints({
            id,
            filter: actualFilter,
            minLevel,
            maxLevel
        });
        
        if (response.status === 202) {
            const data = await DatasourcesService.getEntryPointsFromWebsocket(response.data.link)
            console.log(data)
            return data;
        }
        return Promise.reject()
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