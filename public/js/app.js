createSpinnerContnet= function() {
    return "<div class=\"spinner-wrapper\">\n" +
        "<div class=\"spinner\">\n" +
        "  <div class=\"rect1\"></div>\n" +
        "  <div class=\"rect2\"></div>\n" +
        "  <div class=\"rect3\"></div>\n" +
        "  <div class=\"rect4\"></div>\n" +
        "  <div class=\"rect5\"></div>\n" +
        "</div>" +
        "</div>"
};

loadContent = function(url, container) {
    container.addClass("content-wrapper")
    container.append(createSpinnerContnet());
    $.get(url).done(function(response) {
        container.html(response)
    });
};