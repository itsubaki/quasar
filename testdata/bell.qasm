OPENQASM 3.0;

gate h q { U(pi/2.0, 0, pi) q; gphase(-pi/4); }
gate x q { U(pi, 0, pi) q; gphase(-pi/2); }
gate cx c, t { ctrl @ U(pi, 0, pi) c, t; ctrl @ gphase(-pi/2) c; }

qubit[2] q;
reset q;

h q[0];
cx q[0], q[1];
