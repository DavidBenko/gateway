// Except as noted, this content is licensed under Creative Commons
// Attribution 3.0

/* A little code to ease navigation of these documents.
 *
 * On window load we:
 *  + Generate a table of contents (godocs_generateTOC)
 *  + Add links up to the top of the doc from each section (godocs_addTopLinks)
 */

/* We want to do some stuff on page load (after the HTML is rendered).
   So listen for that:
 */
function bindEvent(el, e, fn) {
  if (el.addEventListener){
    el.addEventListener(e, fn, false);
  } else if (el.attachEvent){
    el.attachEvent('on'+e, fn);
  }
}

function godocs_bindSearchEvents() {
  var search = document.getElementById('search');
  if (!search) {
    // no search box (index disabled)
    return;
  }
  function clearInactive() {
    if (search.className == "inactive") {
      search.value = "";
      search.className = "";
    }
  }
  function restoreInactive() {
    if (search.value !== "") {
      return;
    }
    if (search.type != "search") {
      search.value = search.getAttribute("placeholder");
    }
    search.className = "inactive";
  }
  restoreInactive();
  bindEvent(search, 'focus', clearInactive);
  bindEvent(search, 'blur', restoreInactive);
}

/* Returns the "This sweet header" from <h2>This <i>sweet</i> header</h2>.
 * Takes a node, returns a string.
 */
function godocs_nodeToText(node) {
  var TEXT_NODE = 3; // Defined in Mozilla but not MSIE :(

  var text = '';
  for (var j = 0; j != node.childNodes.length; j++) {
    var child = node.childNodes[j];
    if (child.nodeType == TEXT_NODE) {
      if (child.nodeValue != '[Top]') { //ok, that's a hack, but it works.
        text = text + child.nodeValue;
      }
    } else {
      text = text + godocs_nodeToText(child);
    }
  }
  return text;
}

/* Generates a table of contents: looks for h2 and h3 elements and generates
 * links.  "Decorates" the element with id=="nav" with this table of contents.
 */
function godocs_generateTOC() {
  if (document.getElementById('manual-nav')) { return; }
  var navbar = document.getElementById('nav');
  if (!navbar) { return; }

  var toc_items = [];

  var i;
  var seenNav = false;
  for (i = 0; i < navbar.parentNode.childNodes.length; i++) {
    var node = navbar.parentNode.childNodes[i];
    if (!seenNav) { 
      if (node.id == 'nav') {
        seenNav = true;
      }
      continue;
    }
    if ((node.tagName != 'h2') && (node.tagName != 'H2') &&
        (node.tagName != 'h3') && (node.tagName != 'H3')) {
      continue;
    }
    if (!node.id) {
      node.id = 'tmp_' + i;
    }
    var text = godocs_nodeToText(node);
    if (!text) { continue; }

    var textNode = document.createTextNode(text);

    var link = document.createElement('a');
    link.href = '#' + node.id;
    link.appendChild(textNode);

    // Then create the item itself
    var item;
    if ((node.tagName == 'h2') || (node.tagName == 'H2')) {
      item = document.createElement('dt');
    } else { // h3
      item = document.createElement('dd');
    }

    item.appendChild(link);
    toc_items.push(item);
  }

  if (toc_items.length <= 1) { return; }

  var dl1 = document.createElement('dl');
  var dl2 = document.createElement('dl');

  var split_index = (toc_items.length / 2) + 1;
  if (split_index < 8) {
    split_index = toc_items.length;
  }

  for (i = 0; i < split_index; i++) {
    dl1.appendChild(toc_items[i]);
  }
  for (/* keep using i */; i < toc_items.length; i++) {
    dl2.appendChild(toc_items[i]);
  }

  var tocTable = document.createElement('table');
  navbar.appendChild(tocTable);
  tocTable.className = 'unruled';
  var tocBody = document.createElement('tbody');
  tocTable.appendChild(tocBody);

  var tocRow = document.createElement('tr');
  tocBody.appendChild(tocRow);

  // 1st column
  var tocCell = document.createElement('td');
  tocCell.className = 'first';
  tocRow.appendChild(tocCell);
  tocCell.appendChild(dl1);

  // 2nd column
  tocCell = document.createElement('td');
  tocRow.appendChild(tocCell);
  tocCell.appendChild(dl2);
}

function getElementsByClassName(base, clazz) {
  if (base.getElementsByClassName) {
    return base.getElementsByClassName(clazz);
  }
  var elements = base.getElementsByTagName('*'), foundElements = [];
  for (var n in elements) {
    if (clazz == elements[n].className) {
      foundElements.push(elements[n]);
    }
  }
  return foundElements;
}

function godocs_bindToggle(el) {
  var button = getElementsByClassName(el, "toggleButton");
  var callback = function() {
    if (el.className == "toggle") {
      el.className = "toggleVisible";
    } else {
      el.className = "toggle";
    }
  };
  for (var i = 0; i < button.length; i++) {
    bindEvent(button[i], "click", callback);
  }
}
function godocs_bindToggles(className) {
  var els = getElementsByClassName(document, className);
  for (var i = 0; i < els.length; i++) {
    godocs_bindToggle(els[i]);
  }
}
function godocs_bindToggleLink(l, prefix) {
  bindEvent(l, "click", function() {
    var i = l.href.indexOf("#"+prefix);
    if (i < 0) {
      return;
    }
    var id = prefix + l.href.slice(i+1+prefix.length);
    var eg = document.getElementById(id);
    eg.className = "toggleVisible";
  });
}
function godocs_bindToggleLinks(className, prefix) {
  var links = getElementsByClassName(document, className);
  for (i = 0; i < links.length; i++) {
    godocs_bindToggleLink(links[i], prefix);
  }
}

function godocs_onload() {
  godocs_bindSearchEvents();
  godocs_generateTOC();
  godocs_bindToggles("toggle");
  godocs_bindToggles("toggleVisible");
  godocs_bindToggleLinks("exampleLink", "example_");
  godocs_bindToggleLinks("overviewLink", "");
}

bindEvent(window, 'load', godocs_onload);
