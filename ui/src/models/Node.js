const FastMap = require("collections/fast-map");

export default class Node {

    constructor(name, length, hasContent) {
        this.name = name;
        this.length = length;
        this.hasContent = hasContent;
        this.parent = null;
        this.info = null;
        this.content = null;
        // Entrypoint component (left panel).
        this.component = null;
        // Content component (right panel).
        this.contentComponent = null;

        if (length) {
            this.children = new FastMap();
        } else {
            this.children = null;
        }

        this.level = 0;
    }

    addChildNode(node) {
        if (this.children === null) {
            this.children = new FastMap();
        }

        node.parent = this;
        node.level = this.level + 1;
        this.children.set(node.name, node);
        // TODO Update the length
    }

    hasChildren() {
        return this.children !== null || this.length > 0
    }

    getFullName() {
        if (!this.fullName) {
            this.fullName = (this.parent && this.parent.name ? this.parent.getFullName() + ':' : '') + this.name;
        }
        return this.fullName;
    }

    clear() {
        // eslint-disable-next-line
        console.log("Clearing node %s", this.getFullName());
        this.children = null;
        this.content = null;
    }

}