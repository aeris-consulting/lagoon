export default class JsonHelper {
    /**
    * test if the given string is a valid json string
    */
    static isJson(jsonStringToTest) {
        try {
            JSON.parse(jsonStringToTest);
            return true;
        } catch (error) {
            return false;
        }
    }
}