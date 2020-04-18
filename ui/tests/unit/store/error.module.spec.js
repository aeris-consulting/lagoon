import { mutations, actions } from '../../../src/store/error.module';
import { testAction } from './store.test.helper';
import { REMOVE_ERROR } from '../../../src/store/mutations.type'

describe('mutations', () => {
    it('ADD_ERROR - string error', () => {
        // given
        const state = { errors: [] }
        const errorMessage = 'error message'

        // when
        mutations.ADD_ERROR(state, errorMessage)

        // then
        expect(state.errors.length).toEqual(1)
        expect(state.errors[0]).toEqual({
            message: errorMessage
        })
    })

    it('ADD_ERROR - object type error', () => {
        // given
        const state = { errors: [] }
        const errorMessage = 'error message'
        const error = {
            message: errorMessage
        }

        // when
        mutations.ADD_ERROR(state, error)

        // then
        expect(state.errors.length).toEqual(1)
        expect(state.errors[0]).toEqual({
            message: errorMessage
        })
    })

    it('REMOVE_ERROR', () => {
        // given
        const state = { 
            errors: [
                {
                    message: 'error message'
                }
            ]
        }

        // when
        mutations.REMOVE_ERROR(state, 0)

        // then
        expect(state.errors.length).toEqual(0)
    })
})

describe('actions', () => {
    it('DISSMISS_ERROR', (done) => {
        const state = { errors: [
            {
                message: 'error message'
            }
        ] }

        testAction(actions.DISSMISS_ERROR, 0, state, [
            { type: REMOVE_ERROR, payload: 0}
        ], done)
    })
})
  