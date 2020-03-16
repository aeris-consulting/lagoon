import {
    DISSMISS_ERROR
} from './actions.type';
import {
    ADD_ERROR,
    REMOVE_ERROR
} from './mutations.type';

const initialState = {
    errors: []
}

const state = { ...initialState }

export const actions = {
    [DISSMISS_ERROR](context, errorIndex) {
        context.commit(REMOVE_ERROR, errorIndex);
    }
}

export const mutations = {
    [ADD_ERROR](state, error) {
        if (typeof error === 'string') {
            state.errors.push({
                message: error
            });
        } else {
            state.errors.push(error);
        }
    },
    [REMOVE_ERROR](state, errorIndex) {
        state.errors.splice(errorIndex, 1);
    },
}

export default {
    state,
    actions,
    mutations
}