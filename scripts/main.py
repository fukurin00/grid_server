# coding: utf-8
#!usr/bin/env python

import numpy as np
import yaml
from PIL import Image
import asyncio

import grid as Grid
import mqtt as Mqtt

def on_connect(client, userdata, flag, rc):
    print("Connected with result code " + str(rc))  # 接続できた旨表示
    client.subscribe("/robot/path/1")  # subするトピックを設定
    client.subscribe("/robot/path/2")  # subするトピックを設定

    client.subscribe("/robot/pose/1")  
    client.subscribe("/robot/pose/2")   


# ブローカーが切断したときの処理
def on_disconnect(client, userdata, flag, rc):
    if  rc != 0:
        print("Unexpected disconnection.")

# メッセージが届いたときの処理
def on_message(client, userdata, msg):
    # msg.topicにトピック名が，msg.payloadに届いたデータ本体が入っている
    print("Received message '" + str(msg.payload) + "' on topic '" + msg.topic + "' with QoS " + str(msg.qos))

    topic = str(msg.topic)
    if topic.startswith("/robot/"):
        if topic.startswith("/robot/path"):
            print(msg.payload)
        elif topic.startswith("/robot/pose"):
            print(msg.payload)

def on_publish(client, userdata, mid):     
    print("publish: {0}".format(mid))


def share_path(payload):
    
    pass


def read_map_image(yaml_file_name, map_file_name):
    print('Loading map using ' + yaml_file_name)
    with open(yaml_file_name, "rt") as fp:
        text = fp.read()
        map_info = yaml.safe_load(text)
    im = np.array(Image.open(map_file_name), dtype='int16')
    resolution = map_info['resolution']
    origins = map_info['origin']

    print(im[:,0])

    global ox
    global oy
    ox, oy = [], []
    dx = resolution
    dy = resolution

    insideWall = False
    for (i, line) in enumerate(reversed(im)):
        if i % 2 != 0:
            continue
        for (j, pixel) in enumerate(line):
            if j % 2 != 0:
                continue
            if pixel == 0:
                if insideWall == True:
                    continue
                point = (j * dx + origins[0], i * dy + origins[1])
                ox.append(point[0])
                oy.append(point[1])
            else:
                insideWall = False
    print(len(ox))
    print('Completed loading map')

if __name__ == "__main__":
    yaml_file_name = "../../map/trusco_map_edited.yaml"
    map_file_name = "../../map/trusco_map_edited.pgm"

    read_map_image(yaml_file_name, map_file_name)
    exit()

    global grid
    grid = Grid.Grid(ox=ox, oy=oy, reso=0.5, rr=0.8)

    global waypoints
    waypoints = dict()

    loop = asyncio.get_event_loop()
    
    mqtt = Mqtt.Mqtt(on_connect=on_connect, on_disconnect=on_disconnect, on_message=on_message, on_publish=on_publish)
    loop.run_until_complete(mqtt.run_mqtt())
