export default class FilterHelper {
    static transformFilter(prefix, filter) {
        let actualFilter;
        let overallFilter = ('*' + filter + '*').replace(/[*]+/g, '*');

        if (!prefix) {
            actualFilter = ('*' + filter + '*').replace(/[*]+/g, '*');
        } else {
            let entrypointRegex = new RegExp(prefix);
            prefix = prefix + ':*';
            let overallRegex = new RegExp(overallFilter
                .replace(/^[*]+/g, '')
                .replace(/[*]+$/g, '')
                .replace(/[*]+/g, '.*')
            );

            if (overallRegex.test(prefix)) {
                actualFilter = prefix;
            } else if (entrypointRegex.test(overallFilter
                .replace(/^[*]+/g, '')
                .replace(/[*]+$/g, '')
            )) {
                actualFilter = overallFilter;
            } else {
                actualFilter = overallFilter + ','
                    + prefix
                        .replace(/[*]+/g, '.*')
                        .replace(/[*]+/g, '*');
            }
        }

        return actualFilter;
    }
}
