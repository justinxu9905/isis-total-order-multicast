namespace go multicast.rpc

struct EchoRequest {
    1: string msg;
}

struct EchoResponse {
    1: string msg;
}

struct SendDataRequest {
    1: string msg;
    2: i64 msgId;
    3: string sender;
}

struct SendDataResponse {
    1: i64 msgId;
    2: i64 proposedSeq;
}

struct SendSeqRequest {
    1: i64 msgId;
    2: string sender;
    3: i64 agreedSeq;
    4: string decisionMaker;
}

struct SendSeqResponse {
    1: string deliverTime;
}

service MulticastService {
    EchoResponse echo(1: EchoRequest req);
	SendDataResponse sendData(1: SendDataRequest req);
	SendSeqResponse sendSeq (1: SendSeqRequest req);
}