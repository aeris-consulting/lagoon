import JsonHelper from '../../../src/helpers/JsonHelper.js'

describe('JsonHelper', () => {
  it('testing valid json', () => {
    const validJsonString = '{"prop":123,"arr":["1",5,{"prop":4}]}';
    const isJson = JsonHelper.isJson(validJsonString);
    expect(isJson).toBeTruthy();
  })

  it('testing invalid json', () => {
    const validJsonString = '{"prop":';
    const isJson = JsonHelper.isJson(validJsonString);
    expect(isJson).toBeFalsy();
  })
})
