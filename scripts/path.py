# coding: utf-8
#!usr/bin/env python

import json

class Path:
    def __init__(self, poses, rr, vel):
        self.rr = rr
        self.vel = vel
        self.poses = poses

    def load_path_msg(payload):
        msg = json.open(payload)






class Pos:
    def __init__(self, x,y,z):
        self.x = x
        self.y = y
        self.z = z

class Ori:
    def __init__(self, x, y, z, w)
        self.x = x
        self.y = y
        self.z = z
        self.w = w

class Pose:
    def __init__(pos, ori, est_pass):
        self.pos = pos
        self.ori = oro
        self.passed = False


    