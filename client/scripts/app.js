
var emplogin = ""
var numberEntries;
var prevHash;
var openBox;

//////////// RPC Functions ///////////////
function rpcSend(command, params) {
	var ret = null

	cmd = "EMPService." + command;

	ret = $.ajax({
	    type: "POST",
	    url: "/rpc",
	    headers: {"Authorization" : "Basic " + window.emplogin},
	    // The key needs to match your method's input parameter (case-sensitive).
	    data: JSON.stringify({ method: cmd, params: params, id: 1}),
	    contentType: "application/json; charset=utf-8",
	    dataType: "json",
	    async: false
	});

	if (ret.status != 200) {
		return null
	}

	return $.parseJSON(ret.responseText)
}

function isLoggedIn() {
	if (window.emplogin == "") {
		return false;
	}

	resp = rpcSend("Version", []);
	if (resp == null) {
		return false;
	} else if (resp.error == "Unauthorized") {
		return false;
	}

	return true;
}

function LogIn(user, pass) {
	window.emplogin = window.btoa(user + ":" + pass);
	return isLoggedIn();
}

function addUpdateAddress(formName) {
	var form = document.forms[formName]
	if (form == null) {
		alert("Error: Could not read form.")
		return false
	}

	res = rpcSend("AddUpdateAddress", [{
		address: form["addr"].value,
		address_bytes: null,
		address_label: form["addrlabel"].value,
		registered: form["registered"].checked,
		subscribed: form["subscribed"].checked,
		pubkey: form["pubkey"].value,
		privkey: form["privkey"].value,
		encrypted_privkey: null
	}])

	if (res.error != null) {
		alert("Error Updating Address: " + res.error)
	}

	$.colorbox.close()

	return false
}

function createAddress() {
	res = rpcSend("CreateAddress", [])
	if (res.error != null) {
		alert("Error Creating Address: " + res.error)
	}
	$.colorbox.close()
}

function sendMessage() {
	var form = document.forms["sendmsg"]
	if (form == null) {
		alert("Error: Could not read form.")
		return false
	}

	res = rpcSend("SendMessage", [{
		sender: form["from"].options[form["from"].selectedIndex].value,
		recipient: form["to"].options[form["to"].selectedIndex].value,
		subject: form["subject"].value,
		content: form["message"].value
	}])

	if (res.error != null) {
		alert("Error Sending Message: " + res.error)
	}

	$.colorbox.close()

	return false
}

function delMessage(txidHash) {
	if (confirm("Are you sure you want to delete? This action cannot be undone.")) {
		res = rpcSend("DeleteMessage", [txidHash])
		if (res.error != null) {
			alert("Error Deleting Message: " + res.error)
		}
	}
	$.colorbox.close()
}

function delAddress(addr) {
	if (confirm("Are you sure you want to forget? You will no longer receive messages to this address. This action cannot be undone.")) {
		res = rpcSend("ForgetAddress", [addr])
		if (res.error != null) {
			alert("Error Forgetting Address: " + res.error)
		}
	}
	$.colorbox.close()
}

function purgeMessage(txid) {
	res = rpcSend("PurgeMessage", [txid])

	if (res.error != null) {
		alert("Error Purging Message: " + res.error)
	}

	$.colorbox.close()
}

function pubMessage() {
	var form = document.forms["pubmsg"]
	if (form == null) {
		alert("Error: Could not read form.")
		return false
	}

	res = rpcSend("PublishMessage", [{
		sender: form["from"].options[form["from"].selectedIndex].value,
		subject: form["subject"].value,
		content: form["message"].value
	}])

	if (res.error != null) {
		alert("Error Publishing Message: " + res.error)
	}

	$.colorbox.close()

	return false
}

function credentialCheck() {
	var form = document.forms["loginForm"]
	if (form == null) {
		alert("Error: Could not read form.")
		return false
	}
	var user = form["user"].value
	var pass = form["pass"].value

	if(LogIn(user, pass)) {
		if (form["remember"].checked) {
			setCookie("emplogin", window.emplogin, 1)
		}
		$.colorbox.close();
	} else {
		$("#loginError").show();
	}

	return false
}

//////////// Cookie Functions (from W3Schools) /////////////
function setCookie(cname, cvalue, exdays) {
    var d = new Date();
    d.setTime(d.getTime() + (exdays*24*60*60*1000));
    var expires = "expires="+d.toGMTString();
    document.cookie = cname + "=" + cvalue + "; " + expires;
}

function getCookie(cname) {
    var name = cname + "=";
    var ca = document.cookie.split(';');
    for(var i=0; i<ca.length; i++) {
        var c = ca[i];
        while (c.charAt(0)==' ') c = c.substring(1);
        if (c.indexOf(name) != -1) return c.substring(name.length, c.length);
    }
    return "";
}

/////////////// Util Functions ////////////////////
function ArrayToBase64( buffer ) {
    var binary = ''
    var bytes = new Uint8Array( buffer )
    var len = bytes.byteLength;
    for (var i = 0; i < len; i++) {
        binary += String.fromCharCode( bytes[ i ] )
    }
    return window.btoa( binary );
}

/////////////// Modal Functions ///////////////////

function messageModal(txidHash) {

	res = rpcSend("OpenMessage", [txidHash])
	message = res.result
	date = new Date(Date.parse(message.info.sent));

	var tmp = rpcSend("GetLabel", [message.info.sender]).result
	if (tmp != null) message.info.sender = tmp;
	tmp = rpcSend("GetLabel", [message.info.recipient]).result
	if (tmp != null) message.info.recipient = tmp;

	$("#messageModal").children().children("#sender").text(message.info.sender)
	$("#messageModal").children().children("#recipient").text(message.info.recipient)
	$("#messageModal").children().children("#sent").text(date.toLocaleString())
	if (message.decrypted != null) {
		$("#messageModal").children().children("#subject").text(message.decrypted.Subject)
		$("#messageModal").children().children("#mime").text(message.decrypted.MimeType)
		$("#messageModal").children("#text").text(message.decrypted.Content)
	}

	$("#messageModal").children().children("#purge").attr("onclick", "purgeMessage('" + ArrayToBase64(message.decrypted.Txid) + "')")
	$("#messageModal").children().children("#delete").attr("onclick", "delMessage('" + txidHash + "')")

	$.colorbox({inline:true, href:"#messageModal", width:"50%",
				onLoad:function(){ $("#messageModal").show(); },
				onCleanup:function(){ $("#messageModal").hide(); reloadPage(true); }
				});
}

function newModal() {

	registered = rpcSend("ListAddresses", [true])
	unregistered = rpcSend("ListAddresses", [false])

	$("#newModal").children().children().children("#from").html("")
	$("#newModal").children().children().children("#to").html("")

	for (var i = 0; i < unregistered.result.length; i++) {
		var str
		if (unregistered.result[i][1].length > 0) {
			str = unregistered.result[i][1]
		} else {
			str = unregistered.result[i][0]
		}
		$("#newModal").children().children().children("#to").append("<option value='"+unregistered.result[i][0]+"'>" + str + "</option>")
	}

	for (var i = 0; i < registered.result.length; i++) {
		var str
		if (registered.result[i][1].length > 0) {
			str = registered.result[i][1]
		} else {
			str = registered.result[i][0]
		}
		$("#newModal").children().children().children("#to").append("<option value='"+registered.result[i][0]+"'>" + str + "</option>")
		$("#newModal").children().children().children("#from").append("<option value='"+registered.result[i][0]+"'>" + str + "</option>")
	}



	$.colorbox({inline:true, href:"#newModal", width:"50%",
				onLoad:function(){ $("#newModal").show(); },
				onCleanup:function(){ $("#newModal").hide(); reloadPage(true); }
				});
}

function pubModal() {

	registered = rpcSend("ListAddresses", [true])

	$("#pubModal").children().children().children("#from").html("")

	for (var i = 0; i < registered.result.length; i++) {
		var str
		if (registered.result[i][1].length > 0) {
			str = registered.result[i][1]
		} else {
			str = registered.result[i][0]
		}
		$("#pubModal").children().children().children("#from").append("<option value='"+registered.result[i][0]+"'>" + str + "</option>")
	}



	$.colorbox({inline:true, href:"#pubModal", width:"50%",
				onLoad:function(){ $("#pubModal").show(); },
				onCleanup:function(){ $("#pubModal").hide(); reloadPage(true); }
				});
}

function addrDetailModal(address) {
	addrDetail = rpcSend("GetAddress", [address]).result
	var modal = $("#addrDetailModal")

	modal.children().children("#address").text(addrDetail.address)

	modal.children("form").children("#addr").attr("value", addrDetail.address)
	modal.children("form").children().children("#pubkey").attr("value", addrDetail.public_key)
	modal.children("form").children().children("#privkey").attr("value", addrDetail.private_key)
	document.forms["addrDetail"]["addrlabel"].value = addrDetail.address_label
	document.forms["addrDetail"]["registered"].checked = addrDetail.registered
	document.forms["addrDetail"]["subscribed"].checked = addrDetail.subscribed

	modal.children().children("#forget").attr("onclick", "delAddress('" + addrDetail.address + "')")

	$.colorbox({inline:true, href:"#addrDetailModal", width:"50%",
				onLoad:function(){ $("#addrDetailModal").show(); },
				onCleanup:function(){ $("#addrDetailModal").hide(); reloadPage(true); }
				});
}

function addrModal() {

	openBox = $.colorbox({inline:true, href:"#addrModal", width:"50%",
				onLoad:function(){ $("#addrModal").show(); },
				onCleanup:function(){ $("#addrModal").hide(); reloadPage(true); }
				});
}

function loginModal() {
	$("#loginError").hide();

	$.colorbox({inline:true, href:"#loginModal", width:"50%",
				onLoad:function(){ $("#loginModal").show(); },
				onCleanup:function(){ $("#loginModal").hide(); },
				onClosed:function(){ if(!isLoggedIn()) { loginModal(); } else {reloadPage()}}
				});
}

/////////////// Main Functions //////////////////////
function reloadPage(force) {
	var msg = null
	var addr = null
	var registered

	var status = rpcSend("ConnectionStatus", [])

	switch (status.result) {
		case 1:
		// Client
			$("#status").css("background-color", "yellow");
			break;
		case 2:
		// Full Connection
			$("#status").css("background-color", "green");
			break;
		default:
		// Disconnected
			$("#status").css("background-color", "red");
			break;
	}

	switch (window.location.hash) {
		case "#outbox":
			$("h3#box").text("Outbox");

			msg = rpcSend("Outbox", [])

			break;
		case "#sendbox":
			$("h3#box").text("Sent");
			msg = rpcSend("Sendbox", [])
			break;
		case "#myaddr":
			$("h3#box").text("My Addresses");
			addr = rpcSend("ListAddresses", [true]);
			break;
		case "#address":
			$("h3#box").text("Contacts");
			addr = rpcSend("ListAddresses", [false])
			break;
		case "":
			window.location.hash = "#inbox"
		case "#inbox":
			$("h3#box").text("Inbox");
			msg = rpcSend("Inbox", [])
	}

	$("#refresh").attr("href", window.location.hash)
	var newEntries;

	if (msg == null) newEntries = addr.result.length;
	else newEntries = msg.result.length;

	if (window.location.hash != window.prevHash || newEntries != window.numberEntries || force) {
		window.prevHash = window.location.hash;
		window.numberEntries = newEntries;

		$("table#main").children("colgroup").html("")
		$("table#main").children("thead").html("")
		$("table#main").children("tbody").html("")
		$("table#main").children("tbody").attr("class", "datarow")

		if (msg != null) {
			$("#new").text("New Message")
			$("#new").attr("onclick", "newModal()")
			$("#pub").show();

			$("table#main").attr("class", "table-4")
			for (var i = 0; i < 4; i++) {
				$("table#main").children("colgroup").append("<col span='1'>");
			}
			$("table#main").children("thead").append("\
				<tr>\
	            	<th>Date</th>\
	            	<th>From</th>\
	            	<th>To</th>\
	            	<th>Status</th>\
		        </tr>");
			for (var i = 0; i < msg.result.length; i++) {
				var unread
				if (msg.result[i].read) {
					unread = "Read"
				} else {
					unread = "Unread"
				}

				date = new Date(Date.parse(msg.result[i].sent));

					var tmp = rpcSend("GetLabel", [msg.result[i].sender]).result
					if (tmp == null) {
						if (msg.result[i].sender == null) msg.result[i].sender = "Click to Decrypt..."
						else if (msg.result[i].sender.length < 1) msg.result[i].sender = "Click to Decrypt..."
					} else {
						msg.result[i].sender = tmp
					}

				tmp = rpcSend("GetLabel", [msg.result[i].recipient]).result
				if (tmp == null) {
					if (msg.result[i].recipient == null) {
						msg.result[i].recipient = "&lt;Subscription Message&gt;"
						if (window.location.hash == "#sendbox") unread = "N/A"
					}
				} else {
					msg.result[i].recipient = tmp
				}

				$("table#main").children("tbody").prepend("\
				<tr onclick='messageModal(\"" + ArrayToBase64(msg.result[i].txid_hash) + "\")'>\
	            	<td data-th='Date'>" + date.toLocaleString() + "</td>\
	            	<td data-th='From'>" + msg.result[i].sender + "</td>\
	            	<td data-th='To'>" + msg.result[i].recipient + "</td>\
	            	<td data-th='Status'>" + unread + "</td>\
		        </tr>");
			}
		} else {
			$("#new").text("New Address")
			$("#new").attr("onclick", "addrModal()")
			$("#pub").hide();
			$("table#main").attr("class", "table-2")
			for (var i = 0; i < 2; i++) {
				$("table#main").children("colgroup").append("<col span='1'>");
			}
			$("table#main").children("thead").append("\
				<tr>\
	            	<th>Address</th>\
	            	<th>Label</th>\
		        </tr>");
			for (var i = 0; i < addr.result.length; i++) {
				$("table#main").children("tbody").prepend("\
					<tr onclick='addrDetailModal(\"" + addr.result[i][0] + "\")'>\
						<td data-th='Address'>" + addr.result[i][0] + "</td>\
	            		<td data-th='Label'>" + addr.result[i][1] + "</td>\
	            	</tr>");
			}
		}
	}
}
