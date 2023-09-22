export namespace main {
	
	export class RunParam {
	    mode: string;
	    configName: string;
	    minInterval: number;
	    cycle: number;
	
	    static createFrom(source: any = {}) {
	        return new RunParam(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.mode = source["mode"];
	        this.configName = source["configName"];
	        this.minInterval = source["minInterval"];
	        this.cycle = source["cycle"];
	    }
	}

}

