import math
import collections

class Grid:
    def __init__(self, ox, oy, reso, rr):
        """
        Initialize grid map for a star planning

        ox: x position list of Obstacles [m]
        oy: y position list of Obstacles [m]
        reso: grid resolution [m]
        rr: robot radius[m]
        """

        self.reso = reso
        self.rr = rr
        self.obmap=None
        self.calc_obstacle_map(ox, oy)

        self.rreso = round(rr/(2*reso)) #aspect of robot/grid_size     

    class Node:
        def __init__(self, x, y, cost, pind):
            self.x = x  # index of grid
            self.y = y  # index of grid
            self.cost = cost
            self.pind = pind

        def __str__(self):
            return str(self.x) + "," + str(self.y) + "," + str(
                self.cost) + "," + str(self.pind)


    def calc_obstacle_map(self, ox, oy):
        self.minx = round(min(ox))
        self.miny = round(min(oy))
        self.maxx = round(max(ox))
        self.maxy = round(max(oy))
        print("minx:", self.minx)
        print("miny:", self.miny)
        print("maxx:", self.maxx)
        print("maxy:", self.maxy)

        self.xwidth = int(round((self.maxx - self.minx) / self.reso))
        self.ywidth = int(round((self.maxy - self.miny) / self.reso))
        print("xwidth:", self.xwidth)
        print("ywidth:", self.ywidth)

        # obstacle map generation
        self.obmap = [[False for i in range(self.ywidth)]for i in range(self.xwidth)]
        for ix in range(self.xwidth):
            x = self.calc_xy_grid_position(ix, self.minx)
            for iy in range(self.ywidth):
                y = self.calc_xy_grid_position(iy, self.miny)
                for iox, ioy in zip(ox, oy):
                    d = math.hypot(iox - x, ioy - y)
                    if d <= self.rr:
                        self.obmap[ix][iy] = True
                        break
        print("Completed calculating obstacle_map")

    def calc_xy_grid_position(self, index, minp):
        #xy one direction
        pos = index * self.reso + minp
        return pos

    def calc_grid_position(self, index):
        py = self.minx + round(index // self.xwidth) * self.reso
        px = self.miny + index % self.xwidth * self.reso
        return px,py

    def verify_grid(self, index):
        px, py = self.calc_grid_position(index)
    
        if px < self.minx:
            return False
        elif py < self.miny:
            return False
        elif px >= self.maxx:
            return False
        elif py >= self.maxy:
            return False

        # collision check
        if self.obmap[int(px)][int(py)]:
            return False

        return True

    def calc_grid_index(self, px, py):
        return (py - self.miny) * self.xwidth + (px - self.minx)

    def calc_over_grid_index(self, px, py):
        overs = []
        ci = self.calc_grid_index(px, py)
        overs.append(ci)

        around = [-1,1,-1*self.xwidth, self.xwidth, -1*self.xwidth-1, -1*self.xwidth+1, self.xwidth-1, self.xwidth+1]
        for i in range(self.rreso):
            for v in around:
                ti = ci + (i+1)*v
                if self.verify_grid(ti):
                    tpx, tpy = self.calc_grid_position(ti)
                    d = math.hypot(tpx-px, tpy-py)
                    if d <= self.rr:
                        overs.append(ti)
        return overs


    def check_crush(self, path1:list, path2:list):
        over1 = []
        over2 = []
        for ip1 in path1:
            over1.extend(self.calc_over_grid_index(ip1[0], ip1[1]))
            over1_unique = list(set(over1))
        print(over1_unique)
        for ip2 in path2:
            over2.extend(self.calc_over_grid_index(ip2[0], ip2[1]))
            over2_unique = list(set(over2))
        print(over2_unique)

        #get duplicated path
        over_path = over1_unique + over2_unique
        dup = [k for k, v in collections.Counter(over_path).items() if v>1] 
        return dup


if __name__ == "__main__":
    w1 = {"x":0, "y":0}
    w2 = {"x":1, "y":0}
    w3 = {"x":2, "y":0}
    w4 = {"x":3, "y":0}
    w5 = {"x":4, "y":0}
    w6 = {"x":5, "y":0}
    w7 = {"x":6, "y":0}
    w8 = {"x":7, "y":0}

    path1 = [[0,0],[1,0],[2,0]]
    path2 = [[2,0],[3,0],[4,0]]

    ox = [-1,10]
    oy = [-3,7]

    grid = Grid(ox,oy,0.5,0.25)
    dup = grid.check_crush(path1, path2)
    print(dup)