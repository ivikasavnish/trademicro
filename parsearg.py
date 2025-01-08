import argparse
def parse():
    parser = argparse.ArgumentParser()
    parser.add_argument("symbol", type=str, help="Name of script")
    parser.add_argument("unit", type=int, help="starting quantity")
    parser.add_argument("user", type=str, help="which user")
    parser.add_argument("--diff", type=float, help="percentage gap")
    parser.add_argument("--zag", type=int, help="zag factor",default=1)
    parser.add_argument("--sell", type=int, help="sell")
    parser.add_argument("--storage", type=str, help="storage file")
    parser.add_argument("--config", type=str, help="config")
    parser.add_argument("--type", type=str, help="mis or cnc",default="MTF")
    parser.add_argument("--dd", type=bool,default=True, help="dynamic difference")
    parser.add_argument("--ex", type=str,default="NSE", help="exchange")
    parser.add_argument("--maxlimit",
                        type=int,
                        default=1,
                        help="maximum tolerance")
    parser.add_argument("--preopen", type=bool,default=False, help="preopen")
    args = parser.parse_args()
    return args
