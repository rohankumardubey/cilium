Run the following command to validate that your cluster has proper network
connectivity:

.. code-block:: shell-session

   $ cilium connectivity test
   âšī¸  Monitor aggregation detected, will skip some flow validation steps
   â¨ [k8s-cluster] Creating namespace for connectivity check...
   (...)
   ---------------------------------------------------------------------------------------------------------------------
   đ Test Report
   ---------------------------------------------------------------------------------------------------------------------
   â 69/69 tests successful (0 warnings)

Congratulations! You have a fully functional Kubernetes cluster with Cilium. đ
