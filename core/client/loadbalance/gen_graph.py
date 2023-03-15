import matplotlib.pyplot as plt


def get_data_map() -> dict:
    return {
        "x": [1, 2, 4, 8, 16, 32, 64, 128, 256, 512, 1024, 2048, 4096, 8192, 16384],
        "opt": [10.18, 10.29, 10.11, 10.28, 10.05, 10.40, 10.36, 10.29, 10.34, 10.43, 10.46, 10.37, 10.23, 10.19
            , 11.12],
        "ops": [117657063, 116729284, 117122374, 100000000, 117891826, 119604872,
                111143166, 117566828, 114062922, 117205842, 117551119, 116705707, 111119160, 107241640, 103490913],
    }


if __name__ == '__main__':
    data_map = get_data_map()
    opt_count = 0
    ops_count = 0
    for k, v in data_map.items():
        if k == 'opt':
            for k2, v2 in enumerate(v):
                opt_count += v2
        elif k == 'ops':
            for k2, v2 in enumerate(v):
                ops_count += v2
    print("opt avg -> ", opt_count / len(data_map.get('x')))
    print("ops avg -> ", ops_count / len(data_map.get('x')))
    fig, avs = plt.subplots(2, 1)
    fig.suptitle("Balancer Benchmark, 5000 Node--14C/20T")
    opt_g = avs[0]
    opt_g.set_xlabel("Goroutine Size")
    opt_g.set_ylabel("Option complete time/nano second")
    opt_g.plot(data_map.get('x'), data_map.get('opt'), linestyle='dashdot', color='red', alpha=1)
    ops_g = avs[1]
    ops_g.set_xlabel("Goroutine Size")
    ops_g.set_ylabel("Option per second")
    ops_g.plot(data_map.get('x'), data_map.get('ops'), linestyle='dashdot', color='green', alpha=1)
    plt.show()
