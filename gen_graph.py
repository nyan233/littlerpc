import numpy as np
import matplotlib.pyplot as plt
import sys

pubilc_x = [100, 500, 2000, 5000]
line_color = {
    "grpc": "yellow",
    "arpc": "red",
    "rpcx": "blue",
    "kitex": "black",
    "net/rpc": "green",
    "lrpc": "pink"
}


def print_graph(data_map: dict, tittle: str, style: str) -> None:
    # plt.subplots_adjust(hspace=0.30)
    fig, avs = plt.subplots(2, 2)
    fig.subplots_adjust(hspace=0.07)
    fig.suptitle(tittle)
    layout = ["tps", "p99_latency", "mean_latency", "max_latency"]
    for k, v in enumerate(layout):
        if k == 0:
            gen_sub_graph(avs[0, 0], data_map, style, "tps", "Concurrent Clients", "Thoughts(TPS)")
        elif k == 1:
            gen_sub_graph(avs[0, 1], data_map, style, "p99_latency", "Concurrent Clients", "TP99 Latency(ms)")
        elif k == 2:
            gen_sub_graph(avs[1, 0], data_map, style, "mean_latency", "Concurrent Clients", "Mean Latency(ms)")
        elif k == 3:
            gen_sub_graph(avs[1, 1], data_map, style, "max_latency", "Concurrent Clients", "Max Latency(ms)")
    plt.show()


def gen_sub_graph(avs: object, data_map: dict, pub_style: str, y_key: str, x_label: str, y_label: str) -> None:
    avs.set_xlabel(x_label)
    avs.set_ylabel(y_label)
    for k, v in data_map.items():
        avs.plot(v.get('x', pubilc_x), v.get(y_key), pub_style, color=line_color.get(k), alpha=1, label=k)
    avs.grid(axis='y', color='0.95')
    avs.legend(title='Frameworks')


def get_compute_data_map() -> dict:
    return dict()


def get_echo_data_map() -> dict:
    r0 = {
        "grpc": {
            "x": pubilc_x,
            "tps": [315497, 288126, 195076, 192012],
            "p99_latency": [6379, 1122, 1567, 1268],
            "mean_latency": [133, 164, 244, 247],
            "max_latency": [12587, 2725, 3688, 2853],
        },
        "arpc": {
            "x": pubilc_x,
            "tps": [712859, 722334, 659413, 582648],
            "p99_latency": [369, 317, 502, 956],
            "mean_latency": [67, 67, 73, 82],
            "max_latency": [696, 618, 902, 1424],
        },
        "rpcx": {
            "x": pubilc_x,
            "tps": [553495, 503905, 436528, 395882],
            "p99_latency": [405, 627, 646, 268],
            "mean_latency": [87, 95, 111, 49],
            "max_latency": [687, 1381, 1309, 1064],
        },
        "kitex": {
            "x": pubilc_x,
            "tps": [349101, 317692, 335638, 351790],
            "p99_latency": [759, 1060, 1188, 881],
            "mean_latency": [137, 147, 133, 130],
            "max_latency": [1464, 1860, 1694, 1309],
        },
        "net/rpc": {
            "x": pubilc_x,
            "tps": [366985, 397266, 386443, 370274],
            "p99_latency": [867, 796, 924, 1004],
            "mean_latency": [131, 120, 123, 128],
            "max_latency": [2015, 1084, 1636, 2229],
        },
        "lrpc": {
            "x": pubilc_x,
            "tps": [468274, 529128, 548456, 524136],
            "p99_latency": [515, 456, 329, 379],
            "mean_latency": [99, 89, 88, 92],
            "max_latency": [601, 648, 561, 737],
        },
        "lrpc-async": {
            "x": pubilc_x,
            "tps": [421425, 427240, 429074, 418077],
            "p99_latency": [601, 571, 783, 684],
            "mean_latency": [114, 112, 111, 114],
            "max_latency": [927, 815, 1578, 1345],
        },
        "arpc-async": {
            "x": pubilc_x,
            "tps": [666755, 653295, 641807, 607939],
            "p99_latency": [395, 470, 485, 503],
            "mean_latency": [71, 72, 74, 79],
            "max_latency": [687, 853, 1023, 807],
        }
    }
    return r0


if __name__ == '__main__':
    if len(sys.argv) == 1:
        sys.argv.append("echo")
    print_type = sys.argv[1]
    if print_type == "echo":
        print_graph(get_echo_data_map(), "RPC Mock 10us Benchmark (x=Concurrent Clients,y=$ylabel)", "x--")
    elif print_type == "compute":
        print_graph(get_compute_data_map(), "RPC Compute Benchmark (x=Concurrent Clients,y=$ylabel)", "x--")
