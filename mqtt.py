# coding: utf-8
#!usr/bin/env python

import paho.mqtt.client as mqtt
import time

class Mqtt:
    def __init__(self, on_connect, on_disconnect, on_message, on_publish):
        self.client = mqtt.Client()                 # クラスのインスタンス(実体)の作成
        self.client.on_connect = on_connect         # 接続時のコールバック関数を登録
        self.client.on_disconnect = on_disconnect   # 切断時のコールバックを登録
        self.client.on_message = on_message         # メッセージ到着時のコールバック
        self.client.on_publish = on_publish         # メッセージ送信時のコールバック

        self.client.connect("localhost", 1883, 60)  # 接続先は自分自身          

    def run_mqtt(self):
        self.client.loop_start()    
        while 1:
            self.client.publish("/robot/1/newpath", 1.1)
            time.sleep(1)
    
def on_connect(client, userdata, flag, rc):
    print("Connected with result code " + str(rc))  # 接続できた旨表示
    client.subscribe("/robot/1/path")  # subするトピックを設定
    client.subscribe("/robot/2/path")  # subするトピックを設定

    client.subscribe("/robot/1/pose")  
    client.subscribe("/robot/2/pose")   


# ブローカーが切断したときの処理
def on_disconnect(client, userdata, flag, rc):
    if  rc != 0:
        print("Unexpected disconnection.")

# メッセージが届いたときの処理
def on_message(client, userdata, msg):
    # msg.topicにトピック名が，msg.payloadに届いたデータ本体が入っている
    print("Received message '" + str(msg.payload) + "' on topic '" + msg.topic + "' with QoS " + str(msg.qos))

    if str(msg.topic).startswith("/robot/"):
        print(msg.payload)

def on_publish(client, userdata, mid):     
    print("publish: {0}".format(mid))


if __name__=="__main__":
    # MQTTの接続設定
    mqtt = Mqtt()
    mqtt.run_mqtt()