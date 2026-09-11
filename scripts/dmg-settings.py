"""Finder layout consumed by dmgbuild; coordinates match dmg-background.svg."""
import os
from importlib.metadata import version

if tuple(int(part) for part in version('dmgbuild').split('.')[:3]) < (1, 6, 7):
    raise RuntimeError('dmgbuild 1.6.7 or later is required for macOS Finder backgrounds. See BUILD_AND_INSTALL.md.')

application = defines['app']
assets = defines['assets']
files = [application]
symlinks = {'Applications': '/Applications'}
format = 'UDZO'
filesystem = 'HFS+'
icon = os.path.join(assets, 'installer.icns')
background = os.path.join(assets, 'dmg-background.png')
window_rect = ((160, 160), (720, 420))
icon_locations = {'NET-Test.app': (180, 192), 'Applications': (540, 192)}
default_view = 'icon-view'
show_status_bar = False
show_tab_view = False
show_toolbar = False
show_pathbar = False
show_sidebar = False
show_icon_preview = False
include_icon_view_settings = True
include_list_view_settings = False
arrange_by = None
grid_spacing = 80
scroll_position = (0, 0)
label_pos = 'bottom'
text_size = 16
icon_size = 112
# Setting hide_extensions adds FinderInfo to the app after signing, which
# invalidates strict code-signature verification. Leave the sealed app intact.
hide_extensions = []
