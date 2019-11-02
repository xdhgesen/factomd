import time
import threading
from ctypes import *

class Factomd(threading.Thread):

    def __init__(self, *args, **kwargs):
        self.factomd = cdll.LoadLibrary("./factomd.so")
        self.factomd.Serve.argtypes = []
        self.factomd.Shutdown.argtypes = []
        super(Factomd, self).__init__(*args, **kwargs)

    def run(self):
        self.factomd.Serve()

    def join(self, *args, **kwargs):
        self.factomd.Shutdown()
        super(Factomd, self).join(*args, **kwargs)

node = Factomd()
node.start()
time.sleep(30)
node.join()
