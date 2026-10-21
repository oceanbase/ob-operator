"""Apply an exact, reviewable patch to the signed obdiag 5.0.1 RPM sources."""
from pathlib import Path
import sys


def replace_once(path, before, after):
    text = path.read_text()
    if text.count(before) != 1:
        raise ValueError("unexpected pinned obdiag source: " + str(path))
    path.write_text(text.replace(before, after, 1))


def patch(root):
    replace_once(root / "common/ssh_client/kubernetes_client.py",
        "    def download(self, remote_path, local_path):\n        return self.__download_file_from_pod(self.namespace, self.pod_name, self.container_name, remote_path, local_path)",
        "    def archive_logs(self, source, destination):\n        from src.common.ssh_client.kubernetes_log_transport import archive_logs\n        return archive_logs(self, source, destination)\n\n    def download_log_archive(self, remote_path, local_path):\n        from src.common.ssh_client.kubernetes_log_transport import download\n        return download(self, remote_path, local_path)\n\n    def download(self, remote_path, local_path):\n        return self.__download_file_from_pod(self.namespace, self.pod_name, self.container_name, remote_path, local_path)")
    replace_once(root / "handler/gather/gather_log/base.py",
        '        self.stdio.verbose("gather_log_on_node {0} tar request:{1}".format(self.ssh_client.get_ip(), self.ssh_client.exec_cmd(tar_cmd)))',
        '        if self.ssh_client.ssh_type == "kubernetes":\n            self.ssh_client.client.archive_logs(tmp_log_dir, tar_file)\n        else:\n            self.stdio.verbose("gather_log_on_node {0} tar request:{1}".format(self.ssh_client.get_ip(), self.ssh_client.exec_cmd(tar_cmd)))')
    replace_once(root / "handler/gather/gather_log/base.py",
        '        self.ssh_client.download(tar_file, os.path.join(self.store_dir, os.path.basename("{0}".format(tar_file))))',
        '        destination = os.path.join(self.store_dir, os.path.basename(tar_file))\n        if self.ssh_client.ssh_type == "kubernetes":\n            self.ssh_client.client.download_log_archive(tar_file, destination)\n        else:\n            self.ssh_client.download(tar_file, destination)')


if __name__ == "__main__":
    patch(Path(sys.argv[1]))
