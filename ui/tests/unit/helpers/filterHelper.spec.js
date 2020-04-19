import FilterHelper from '../../../src/helpers/filterHelper'

describe('FilterHelper', () => {
  it('no prefix', () => {
    // give
    const prefix = ''
    const filter = 'foo'

    // when
    const rs = FilterHelper.transformFilter(prefix, filter)

    // then
    expect(rs).toEqual('*foo*')
  })

  it('no filter', () => {
    // give
    const prefix = 'foo'
    const filter = ''

    // when
    const rs = FilterHelper.transformFilter(prefix, filter)

    // then
    expect(rs).toEqual('foo:*')
  })

  it('prefix and filter', () => {
    // give
    const prefix = 'bar'
    const filter = 'foo'

    // when
    const rs = FilterHelper.transformFilter(prefix, filter)

    // then
    expect(rs).toEqual('*foo*,bar:.*')
  })
})
