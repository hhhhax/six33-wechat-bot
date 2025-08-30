export namespace backend {
	
	export class BetDetail {
	    numbers: number[];
	    amount: number;
	    groups: number;
	    description: string;
	
	    static createFrom(source: any = {}) {
	        return new BetDetail(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.numbers = source["numbers"];
	        this.amount = source["amount"];
	        this.groups = source["groups"];
	        this.description = source["description"];
	    }
	}
	export class BetTypeSummary {
	    type: string;
	    lottery: string;
	    total_groups: number;
	    total_amount: number;
	
	    static createFrom(source: any = {}) {
	        return new BetTypeSummary(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.type = source["type"];
	        this.lottery = source["lottery"];
	        this.total_groups = source["total_groups"];
	        this.total_amount = source["total_amount"];
	    }
	}
	export class ParsedBet {
	    type: string;
	    lottery: string;
	    numbers: number[];
	    amount: number;
	    total_amount: number;
	    groups: number;
	    description: string;
	    original: string;
	    is_error: boolean;
	
	    static createFrom(source: any = {}) {
	        return new ParsedBet(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.type = source["type"];
	        this.lottery = source["lottery"];
	        this.numbers = source["numbers"];
	        this.amount = source["amount"];
	        this.total_amount = source["total_amount"];
	        this.groups = source["groups"];
	        this.description = source["description"];
	        this.original = source["original"];
	        this.is_error = source["is_error"];
	    }
	}
	export class BetParseResponse {
	    success: boolean;
	    error: string;
	    results: ParsedBet[];
	    total_bets: number;
	    total_amount: number;
	    total_groups: number;
	    summary: BetTypeSummary[];
	    parse_time: string;
	
	    static createFrom(source: any = {}) {
	        return new BetParseResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.error = source["error"];
	        this.results = this.convertValues(source["results"], ParsedBet);
	        this.total_bets = source["total_bets"];
	        this.total_amount = source["total_amount"];
	        this.total_groups = source["total_groups"];
	        this.summary = this.convertValues(source["summary"], BetTypeSummary);
	        this.parse_time = source["parse_time"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class BetTypeStat {
	    amount: number;
	    groups: number;
	    count: number;
	
	    static createFrom(source: any = {}) {
	        return new BetTypeStat(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.amount = source["amount"];
	        this.groups = source["groups"];
	        this.count = source["count"];
	    }
	}
	export class RoundBetStatistics {
	    totalAmount: number;
	    totalGroups: number;
	    totalBets: number;
	    lotteryBetTypeStats: Record<string, any>;
	    lotteryTotals: Record<string, BetTypeStat>;
	    betTypeTotals: Record<string, BetTypeStat>;
	
	    static createFrom(source: any = {}) {
	        return new RoundBetStatistics(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.totalAmount = source["totalAmount"];
	        this.totalGroups = source["totalGroups"];
	        this.totalBets = source["totalBets"];
	        this.lotteryBetTypeStats = source["lotteryBetTypeStats"];
	        this.lotteryTotals = this.convertValues(source["lotteryTotals"], BetTypeStat, true);
	        this.betTypeTotals = this.convertValues(source["betTypeTotals"], BetTypeStat, true);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class BetStatistics {
	    totalAmount: number;
	    totalGroups: number;
	    lotteryCount: number;
	    lotteryBetTypeStats: Record<string, any>;
	
	    static createFrom(source: any = {}) {
	        return new BetStatistics(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.totalAmount = source["totalAmount"];
	        this.totalGroups = source["totalGroups"];
	        this.lotteryCount = source["lotteryCount"];
	        this.lotteryBetTypeStats = source["lotteryBetTypeStats"];
	    }
	}
	export class BetTypeInfo {
	    betType: string;
	    betDetails: BetDetail[];
	    totalGroups: number;
	    totalAmount: number;
	    isComplex: boolean;
	    isDrag: boolean;
	    hasNumbers: boolean;
	
	    static createFrom(source: any = {}) {
	        return new BetTypeInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.betType = source["betType"];
	        this.betDetails = this.convertValues(source["betDetails"], BetDetail);
	        this.totalGroups = source["totalGroups"];
	        this.totalAmount = source["totalAmount"];
	        this.isComplex = source["isComplex"];
	        this.isDrag = source["isDrag"];
	        this.hasNumbers = source["hasNumbers"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class LotteryBetInfo {
	    lotteryType: string;
	    betTypes: Record<string, BetTypeInfo>;
	    totalAmount: number;
	    totalGroups: number;
	
	    static createFrom(source: any = {}) {
	        return new LotteryBetInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.lotteryType = source["lotteryType"];
	        this.betTypes = this.convertValues(source["betTypes"], BetTypeInfo, true);
	        this.totalAmount = source["totalAmount"];
	        this.totalGroups = source["totalGroups"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class SingleBetParsing {
	    betId: string;
	    originalText: string;
	    lotteryBets: Record<string, LotteryBetInfo>;
	    betStatistics: BetStatistics;
	    hasError: boolean;
	    errorMessage: string;
	
	    static createFrom(source: any = {}) {
	        return new SingleBetParsing(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.betId = source["betId"];
	        this.originalText = source["originalText"];
	        this.lotteryBets = this.convertValues(source["lotteryBets"], LotteryBetInfo, true);
	        this.betStatistics = this.convertValues(source["betStatistics"], BetStatistics);
	        this.hasError = source["hasError"];
	        this.errorMessage = source["errorMessage"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class BetParsingResult {
	    roundId: string;
	    originalText: string;
	    parsedBets: SingleBetParsing[];
	    roundStatistics: RoundBetStatistics;
	    // Go type: time
	    parseTime: any;
	    hasError: boolean;
	    errorMessages: string[];
	
	    static createFrom(source: any = {}) {
	        return new BetParsingResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.roundId = source["roundId"];
	        this.originalText = source["originalText"];
	        this.parsedBets = this.convertValues(source["parsedBets"], SingleBetParsing);
	        this.roundStatistics = this.convertValues(source["roundStatistics"], RoundBetStatistics);
	        this.parseTime = this.convertValues(source["parseTime"], null);
	        this.hasError = source["hasError"];
	        this.errorMessages = source["errorMessages"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	export class BetTypeAliases {
	    three_of_three: string[];
	    three_of_two: string[];
	    two_of_two: string[];
	    special: string[];
	
	    static createFrom(source: any = {}) {
	        return new BetTypeAliases(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.three_of_three = source["three_of_three"];
	        this.three_of_two = source["three_of_two"];
	        this.two_of_two = source["two_of_two"];
	        this.special = source["special"];
	    }
	}
	
	
	
	export class ColorConfig {
	    red: number[];
	    green: number[];
	    blue: number[];
	
	    static createFrom(source: any = {}) {
	        return new ColorConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.red = source["red"];
	        this.green = source["green"];
	        this.blue = source["blue"];
	    }
	}
	export class HitThreeOdds {
	    odds_ratio: number;
	    rebate: number;
	
	    static createFrom(source: any = {}) {
	        return new HitThreeOdds(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.odds_ratio = source["odds_ratio"];
	        this.rebate = source["rebate"];
	    }
	}
	export class HitTwoOdds {
	    odds_ratio: number;
	    rebate: number;
	
	    static createFrom(source: any = {}) {
	        return new HitTwoOdds(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.odds_ratio = source["odds_ratio"];
	        this.rebate = source["rebate"];
	    }
	}
	export class KeywordAliases {
	    new_macau: string[];
	    old_macau: string[];
	    hong_kong: string[];
	    complex: string[];
	    each: string[];
	    per_group: string[];
	
	    static createFrom(source: any = {}) {
	        return new KeywordAliases(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.new_macau = source["new_macau"];
	        this.old_macau = source["old_macau"];
	        this.hong_kong = source["hong_kong"];
	        this.complex = source["complex"];
	        this.each = source["each"];
	        this.per_group = source["per_group"];
	    }
	}
	
	export class SpecialOdds {
	    odds_ratio: number;
	    rebate: number;
	
	    static createFrom(source: any = {}) {
	        return new SpecialOdds(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.odds_ratio = source["odds_ratio"];
	        this.rebate = source["rebate"];
	    }
	}
	export class TwoOfTwoOdds {
	    odds_ratio: number;
	    rebate: number;
	
	    static createFrom(source: any = {}) {
	        return new TwoOfTwoOdds(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.odds_ratio = source["odds_ratio"];
	        this.rebate = source["rebate"];
	    }
	}
	export class ThreeOfTwoOdds {
	    hit_two_odds: HitTwoOdds;
	    hit_three_odds: HitThreeOdds;
	
	    static createFrom(source: any = {}) {
	        return new ThreeOfTwoOdds(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.hit_two_odds = this.convertValues(source["hit_two_odds"], HitTwoOdds);
	        this.hit_three_odds = this.convertValues(source["hit_three_odds"], HitThreeOdds);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class ThreeOfThreeOdds {
	    odds_ratio: number;
	    rebate: number;
	
	    static createFrom(source: any = {}) {
	        return new ThreeOfThreeOdds(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.odds_ratio = source["odds_ratio"];
	        this.rebate = source["rebate"];
	    }
	}
	export class OddsConfig {
	    three_of_three: ThreeOfThreeOdds;
	    three_of_two: ThreeOfTwoOdds;
	    two_of_two: TwoOfTwoOdds;
	    special: SpecialOdds;
	
	    static createFrom(source: any = {}) {
	        return new OddsConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.three_of_three = this.convertValues(source["three_of_three"], ThreeOfThreeOdds);
	        this.three_of_two = this.convertValues(source["three_of_two"], ThreeOfTwoOdds);
	        this.two_of_two = this.convertValues(source["two_of_two"], TwoOfTwoOdds);
	        this.special = this.convertValues(source["special"], SpecialOdds);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	
	
	
	export class TailConfig {
	    tail_0: number[];
	    tail_1: number[];
	    tail_2: number[];
	    tail_3: number[];
	    tail_4: number[];
	    tail_5: number[];
	    tail_6: number[];
	    tail_7: number[];
	    tail_8: number[];
	    tail_9: number[];
	
	    static createFrom(source: any = {}) {
	        return new TailConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.tail_0 = source["tail_0"];
	        this.tail_1 = source["tail_1"];
	        this.tail_2 = source["tail_2"];
	        this.tail_3 = source["tail_3"];
	        this.tail_4 = source["tail_4"];
	        this.tail_5 = source["tail_5"];
	        this.tail_6 = source["tail_6"];
	        this.tail_7 = source["tail_7"];
	        this.tail_8 = source["tail_8"];
	        this.tail_9 = source["tail_9"];
	    }
	}
	export class ZodiacConfig {
	    rat: number[];
	    ox: number[];
	    tiger: number[];
	    rabbit: number[];
	    dragon: number[];
	    snake: number[];
	    horse: number[];
	    goat: number[];
	    monkey: number[];
	    rooster: number[];
	    dog: number[];
	    pig: number[];
	
	    static createFrom(source: any = {}) {
	        return new ZodiacConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.rat = source["rat"];
	        this.ox = source["ox"];
	        this.tiger = source["tiger"];
	        this.rabbit = source["rabbit"];
	        this.dragon = source["dragon"];
	        this.snake = source["snake"];
	        this.horse = source["horse"];
	        this.goat = source["goat"];
	        this.monkey = source["monkey"];
	        this.rooster = source["rooster"];
	        this.dog = source["dog"];
	        this.pig = source["pig"];
	    }
	}
	export class SystemConfig {
	    zodiac_config: ZodiacConfig;
	    color_config: ColorConfig;
	    tail_config: TailConfig;
	    bet_type_aliases: BetTypeAliases;
	    keyword_aliases: KeywordAliases;
	    odds_config: OddsConfig;
	
	    static createFrom(source: any = {}) {
	        return new SystemConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.zodiac_config = this.convertValues(source["zodiac_config"], ZodiacConfig);
	        this.color_config = this.convertValues(source["color_config"], ColorConfig);
	        this.tail_config = this.convertValues(source["tail_config"], TailConfig);
	        this.bet_type_aliases = this.convertValues(source["bet_type_aliases"], BetTypeAliases);
	        this.keyword_aliases = this.convertValues(source["keyword_aliases"], KeywordAliases);
	        this.odds_config = this.convertValues(source["odds_config"], OddsConfig);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	
	
	

}

