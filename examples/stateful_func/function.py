import redis

def handler(params, context):
    try:
        key = params["n"]
        val = ''.join(fibonacci_nums(int(key)))
        result = set_on_redis(key, val)
        result=True
        if result:
            return {"Fibonacci": "[{key}:{val}], added on Redis!".format(key = key,val=val) }
        else:
             return {}
    except:
        return {}



def fibonacci_nums(n):
    sequence = ""
    if n <= 0:
        sequence += "0"
        return sequence
    sequence = "0, 1"
    count = 2
    n1 = 0
    n2 = 1
    while count <= n:
        next_value = n2 + n1
        sequence += "," + "".join(str(next_value))
        n1 = n2
        n2 = next_value
        count += 1
    return sequence


#connection to redis server and create/update (if key already exists) the pair {key:value}
def set_on_redis(key, value):
    r = redis.Redis(host="192.168.122.2", port=6379, db=0)
    try:
        r.set(key, value)
    except redis.RedisError as re:
        print(re)
        return False
    return True

